# MicroEngine Documentation

## Overview

The MicroEngine is a highly configurable rule-based matching engine for filtering network requests based on various criteria including IP addresses, URLs, and request paths. It supports complex rule definitions with flexible condition combinations and efficient matching algorithms.

## Architecture

### Core Components

1. **RuleEngine**: The central component that manages rules and IP groups, handling rule matching and evaluation.
2. **Rule**: Represents a security rule with matching conditions and actions (blacklist/whitelist).
3. **Condition System**: A flexible system that supports both simple conditions and complex composite conditions.
4. **Matcher Interface**: Defines the contract for condition matching, implemented by different condition types.

### Condition Types

1. **SimpleCondition**: Basic condition that matches against a specific target (IP, URL, path) with various matching strategies.
2. **CompositeCondition**: Complex condition that combines multiple conditions using logical operators (AND/OR).

## Core Execution Flow

### Workflow

1. **Initialization**:
   - Create a new RuleEngine instance
   - Configure MongoDB connection (if using database storage)
   - Load IP groups and rules from MongoDB or JSON

2. **Rule Loading Process**:
   - Load rules with priority and sequence information
   - Parse conditions from BSON/JSON to condition objects
   - Sort rules by priority (higher values first) and sequence number

3. **Request Matching Flow**:
   - Validate the incoming request data (IP, URL, path)
   - Iterate through rules in priority order
   - Evaluate each rule's conditions against the request
   - Return match result based on rule type (blacklist/whitelist)
   - Default behavior: block if whitelist rules exist but none match, otherwise allow

4. **Condition Matching**:
   - Simple conditions: Direct matching against targets
   - Composite conditions: Recursive evaluation of nested conditions with logical operators
   - Short-circuit evaluation for performance optimization (early exit when result determined)

### Flow Diagram (Version 1) - Business Process Focus

```
                           ┌───────────────────┐
                           │  Incoming Request │
                           └─────────┬─────────┘
                                     ▼
┌─────────────────────────────────────────────────────────────┐
│                       Rule Engine                            │
│                                                             │
│  ┌─────────────────┐    ┌─────────────────┐                 │
│  │   Rule Loading  │    │   IP Groups     │                 │
│  │                 │    │                 │                 │
│  │ - Load from DB  │    │ - IP Collections│                 │
│  │ - Sort by       │    │ - CIDR Ranges   │                 │
│  │   priority      │    │                 │                 │
│  └────────┬────────┘    └─────────────────┘                 │
│           │                                                 │
│           ▼                                                 │
│  ┌────────────────────────────────────────┐                 │
│  │            Rule Processing              │                 │
│  │                                         │                 │
│  │  ┌───────────┐   ┌────────────────┐    │                 │
│  │  │Simple Rule│   │ Composite Rule │    │                 │
│  │  │Evaluation │   │   Evaluation   │    │                 │
│  │  └───────────┘   └────────────────┘    │                 │
│  │                                         │                 │
│  └──────────────────┬─────────────────────┘                 │
│                     │                                       │
│                     ▼                                       │
│  ┌────────────────────────────────────────┐                 │
│  │         Decision Application            │                 │
│  │                                         │                 │
│  │  ┌────────────┐    ┌────────────┐      │                 │
│  │  │  Blacklist │    │  Whitelist │      │                 │
│  │  │   Action   │    │   Action   │      │                 │
│  │  └────────────┘    └────────────┘      │                 │
│  │                                         │                 │
│  └──────────────────┬─────────────────────┘                 │
│                     │                                       │
└─────────────────────┼───────────────────────────────────────┘
                      ▼
             ┌─────────────────┐
             │  Allow/Block    │
             │    Decision     │
             └─────────────────┘
```

### Flow Diagram (Version 2) - Implementation Details Focus

```
┌───────────────────────────────────────────────────────────────────────────────┐
│                             INITIALIZATION                                     │
├───────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  NewRuleEngine() → RuleEngine{                                                │
│     Rules: make([]Rule, 0),                                                   │
│     IPGroups: make(map[string]*model.IPGroup),                                │
│     regexCache: make(map[string]*regexp.Regexp),                              │
│  }                                                                            │
│                                                                               │
│  InitMongoConfig(config *MongoDBConfig) → Set e.mongoConfig                   │
│                                                                               │
│  LoadAllFromMongoDB() → {                                                     │
│     LoadIPGroupsFromMongoDB()                                                 │
│     LoadRulesFromMongoDB()                                                    │
│  }                                                                            │
│                                                                               │
└───────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌───────────────────────────────────────────────────────────────────────────────┐
│                           RULE LOADING & PARSING                               │
├───────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  LoadRulesFromMongoDB() → {                                                   │
│     1. Get rules collection                                                   │
│     2. Check/create default rule if needed                                    │
│     3. Query all rules: cursor, err := collection.Find(ctx, bson.D{})         │
│     4. Parse rules: cursor.All(ctx, &rules)                                   │
│     5. Set sequence: rules[i].sequence = i                                    │
│     6. Parse conditions:                                                      │
│        for i := range rules {                                                 │
│          parsedCondition, err := e.factory.ParseCondition(rule.Condition)     │
│          rule.parsedCondition = parsedCondition                               │
│        }                                                                      │
│     7. Sort by priority and sequence                                          │
│  }                                                                            │
│                                                                               │
│  ParseCondition(data bson.Raw) → Matcher {                                    │
│     1. Unmarshal condition type                                               │
│     2. If SimpleConditionType:                                                │
│        - Return SimpleCondition                                               │
│     3. If CompositeConditionType:                                             │
│        - Parse all nested conditions recursively                              │
│        - Store in condition.parsedConditions                                  │
│  }                                                                            │
│                                                                               │
└───────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌───────────────────────────────────────────────────────────────────────────────┐
│                             REQUEST MATCHING                                   │
├───────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  MatchRequest(ip, url, path) → (bool, model.RuleType, *Rule, error) {         │
│     1. Validate IP: if !isValidIP(ip) { return error }                        │
│     2. Check for whitelists: hasWhitelistRule = any rule is whitelist         │
│     3. Iterate through rules (already sorted):                                │
│        - Skip disabled rules: if r.Status == model.RuleDisabled { continue }  │
│        - Test rule: match, err := r.parsedCondition.Match(e, ip, url, path)   │
│        - If matched blacklist: return true (block), r.Type, &r, nil           │
│        - If matched whitelist: return false (allow), r.Type, &r, nil          │
│     4. Default behavior:                                                      │
│        - If whitelist exists but none matched: return true (block)            │
│        - Otherwise: return false (allow)                                      │
│  }                                                                            │
│                                                                               │
└───────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌───────────────────────────────────────────────────────────────────────────────┐
│                           CONDITION MATCHING                                   │
├───────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  SimpleCondition.Match(eng, ip, url, path) → (bool, error) {                  │
│     - Based on Target:                                                        │
│       - SourceIP: return eng.matchIP(c, ip)                                   │
│       - TargetURL: return eng.matchURL(c, url)                                │
│       - TargetPath: return eng.matchPath(c, path)                             │
│  }                                                                            │
│                                                                               │
│  matchIP/matchURL/matchPath methods apply specific matchers:                  │
│     - MatchEqual: direct comparison                                           │
│     - MatchInIPGroup: check all IPs in group                                  │
│     - MatchRegex: use cached regex patterns                                   │
│     - etc.                                                                    │
│                                                                               │
│  CompositeCondition.Match(eng, ip, url, path) → (bool, error) {               │
│     1. Start with default result based on operator:                           │
│        - AND: result = true                                                   │
│        - OR: result = false                                                   │
│     2. Iterate through nested conditions:                                     │
│        - Recursively call Match() on each condition                           │
│        - Apply short-circuit logic:                                           │
│          - AND: if !match return false (short-circuit)                        │
│          - OR: if match return true (short-circuit)                           │
│     3. Return final result                                                    │
│  }                                                                            │
│                                                                               │
└───────────────────────────────────────────────────────────────────────────────┘
```

## Recursion Flow in Composite Conditions

The composite condition evaluation uses recursion to process nested conditions:

1. The `ParseCondition` method recursively parses nested conditions in composite conditions
2. During matching, `CompositeCondition.Match()` recursively calls the `Match()` method on all child conditions
3. Logical operations (AND/OR) are applied with short-circuit evaluation:
   - For AND: Return false when any condition fails
   - For OR: Return true when any condition succeeds

### Condition Factory and Recursive Parsing Flow Diagram

The rule factory enables arbitrary condition combinations through recursive parsing. This is a key feature that allows for creating complex rule hierarchies with unlimited nesting.

```
┌───────────────────────────────────────────────────────────────────────────────┐
│                       CONDITION FACTORY & PARSING                              │
├───────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  ParseCondition(data bson.Raw)                                                │
│  ┌─────────────────────────────┐                                              │
│  │ Extract Base Condition Type │                                              │
│  └──────────────┬──────────────┘                                              │
│                 │                                                             │
│                 ▼                                                             │
│  ┌─────────────────────────────┐                                              │
│  │ Switch on Condition Type    │                                              │
│  └──────────────┬──────────────┘                                              │
│                 │                                                             │
│        ┌────────┴────────┐                                                    │
│        │                 │                                                    │
│        ▼                 ▼                                                    │
│  ┌────────────┐   ┌────────────────────────────┐                              │
│  │ Simple     │   │ Composite                  │                              │
│  │ Condition  │   │ Condition                  │                              │
│  └─────┬──────┘   └─────────────┬──────────────┘                              │
│        │                        │                                             │
│        │                        ▼                                             │
│        │          ┌────────────────────────────┐                              │
│        │          │ Initialize Empty Array     │◄─────┐                       │
│        │          │ condition.parsedConditions │                              │
│        │          └─────────────┬──────────────┘                              │
│        │                        │                                             │
│        │                        ▼                                             │
│        │          ┌────────────────────────────┐                              │
│        │          │ For Each Child Condition   │◄─────┐                       │
│        │          └─────────────┬──────────────┘      │                       │
│        │                        │                     │                       │
│        │                        ▼                     │                       │
│        │          ┌────────────────────────────┐      │                       │
│        │          │ RECURSIVE CALL:            │      │                       │
│        │          │ ParseCondition(childData)  │      │                       │
│        │          └─────────────┬──────────────┘      │                       │
│        │                        │                     │                       │
│        │                        ▼                     │                       │
│        │          ┌────────────────────────────┐      │                       │
│        │          │ Append Child to            │      │                       │
│        │          │ parsedConditions Array     │──────┘                       │
│        │          └────────────────────────────┘                              │
│        │                                                                      │
│        └───────────────┐ ┌───────────────────────                             │
│                        │ │                                                    │
│                        ▼ ▼                                                    │
│               ┌────────────────────┐                                          │
│               │ Return Condition   │                                          │
│               │ Object Implementing│                                          │
│               │ Matcher Interface  │                                          │
│               └────────────────────┘                                          │
│                                                                               │
└───────────────────────────────────────────────────────────────────────────────┘
```

### Condition Evaluation and Tree Traversal Flow Diagram

When evaluating a request against the rules, the condition tree is traversed recursively:

```
┌───────────────────────────────────────────────────────────────────────────────┐
│                   CONDITION EVALUATION (TREE TRAVERSAL)                        │
├───────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  Rule Evaluation Process                                                      │
│  ┌──────────────────────────┐                                                 │
│  │ rule.parsedCondition.Match(engine, ip, url, path)                          │
│  └────────────────┬─────────┘                                                 │
│                   │                                                           │
│                   ▼                                                           │
│  ┌─────────────────────────────┐                                              │
│  │ Is SimpleCondition?         │                                              │
│  └────────────┬────────────────┘                                              │
│               │                                                               │
│      ┌────────┴─────────┐                                                     │
│      │                  │                                                     │
│      ▼                  ▼                                                     │
│  ┌────────────┐  ┌─────────────────┐                                          │
│  │   Simple   │  │   Composite     │                                          │
│  │  Matching  │  │   Matching      │                                          │
│  └─────┬──────┘  └────────┬────────┘                                          │
│        │                  │                                                   │
│        ▼                  ▼                                                   │
│  ┌────────────┐  ┌─────────────────────────┐                                  │
│  │ Match on   │  │ Initialize result based │                                  │
│  │ Target Type│  │ on operator (AND/OR)    │                                  │
│  │ - IP       │  └────────────┬────────────┘                                  │
│  │ - URL      │               │                                               │
│  │ - Path     │               ▼                                               │
│  └─────┬──────┘  ┌─────────────────────────┐                                  │
│        │         │ For each child condition│◄────┐                            │
│        │         └────────────┬────────────┘     │                            │
│        │                      │                  │                            │
│        │                      ▼                  │                            │
│        │         ┌─────────────────────────┐     │                            │
│        │         │ RECURSIVE CALL:         │     │                            │
│        │         │ condition.Match()       │     │                            │
│        │         └────────────┬────────────┘     │                            │
│        │                      │                  │                            │
│        │                      ▼                  │                            │
│        │         ┌─────────────────────────┐     │                            │
│        │         │ Apply short-circuit     │     │                            │
│        │         │ logic:                  │     │                            │
│        │         │ - AND: if !match return │     │                            │
│        │         │   false immediately     │     │                            │
│        │         │ - OR: if match return   │     │                            │
│        │         │   true immediately      │     │                            │
│        │         └────────────┬────────────┘     │                            │
│        │                      │                  │                            │
│        │                      │                  │                            │
│        │                      │ (if no short-circuit)                         │
│        │                      ▼                  │                            │
│        │         ┌─────────────────────────┐     │                            │
│        │         │ Continue to next child  │─────┘                            │
│        │         └─────────────────────────┘                                  │
│        │                                                                      │
│        └───────────────┐ ┌───────────────────────                             │
│                        │ │                                                    │
│                        ▼ ▼                                                    │
│               ┌────────────────────┐                                          │
│               │ Return Final Match │                                          │
│               │ Result (bool)      │                                          │
│               └────────────────────┘                                          │
│                                                                               │
└───────────────────────────────────────────────────────────────────────────────┘
```

This recursive structure enables complex condition trees of arbitrary depth and complexity, such as:

```
                       Root Condition (AND)
                      /                   \
                     /                     \
            Simple Condition        Composite Condition (OR)
           (IP in blacklist)        /                      \
                                   /                        \
                       Simple Condition           Composite Condition (AND)
                      (URL contains "/admin")    /                        \
                                               /                          \
                                    Simple Condition            Simple Condition
                                  (Path starts with "/api")   (URL contains ".php")
```

## Design Principles

1. **Flexibility**: Support for various matching strategies and complex condition combinations
2. **Performance Optimization**:
   - Regular expression caching
   - Short-circuit evaluation for logical operators
   - Priority-based rule processing to optimize common cases
   - Efficient IP matching algorithms (IP groups, CIDR matching)

3. **Extensibility**:
   - Interface-based design allows adding new condition types
   - Factory pattern for condition creation
   - Clear separation of concerns between rule storage and rule evaluation

4. **Resilience**:
   - Comprehensive error handling
   - Validation of inputs (IP addresses, CIDR ranges, regex patterns)
   - Default rules for system protection

## Performance Considerations

1. **RegEx Caching**: The engine maintains a cache of compiled regular expressions to avoid recompilation
2. **Rule Prioritization**: Rules are sorted by priority to ensure important rules are checked first
3. **Short-circuit Evaluation**: Logical operations stop as soon as the result is determined
4. **CIDR Optimization**: IP addresses are efficiently matched against CIDR ranges
5. **Future Optimizations** (TODOs in code):
   - Replace linear scanning of IP groups with Radix Tree (Patricia Trie)
   - Implement LRU caching for regular expressions
   - Add expiration to cached items to prevent cache growth

## Maintenance Guidelines

1. **Adding New Match Types**:
   - Add a new constant to the appropriate MatchType enum
   - Implement matching logic in the corresponding match function
   - Update validation in the related condition type

2. **Adding New Target Types**:
   - Add a new constant to the TargetType enum
   - Add case handling in the Match method of SimpleCondition
   - Implement a new matching function in the RuleEngine

3. **Performance Tuning**:
   - Monitor and optimize regexCache size
   - Review rule priority assignments for optimal processing order
   - Consider implementing the TODOs marked in the code for additional optimization 