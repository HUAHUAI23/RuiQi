# ConditionBuilder Component Documentation

## Overview

The `ConditionBuilder` component is a recursive React component used for creating and visualizing complex nested condition hierarchies in the WAF interface. It provides an intuitive UI for building both simple conditions and composite conditions with arbitrary nesting depth.

## Component Architecture

The component implements a self-recursive pattern to render condition trees, matching the backend rule engine's condition model.

### Core Features

1. **Recursive Rendering**: Can render itself as a child component, creating nested condition groups
2. **Two Condition Types**: Supports simple conditions (leaf nodes) and composite conditions (container nodes)
3. **Dynamic Condition Management**: Allows adding, removing, and configuring conditions through the UI
4. **Logical Operators**: Supports AND/OR operators with visual indicators
5. **Visual Connectivity**: Displays connecting lines to visually represent the logical hierarchy

## Recursive UI Structure

```
ConditionBuilder (root - composite condition)
├── Operator badge (AND/OR)
├── Child conditions (array)
│   ├── ConditionBuilder (simple condition)
│   ├── ConditionBuilder (composite condition)
│   │   ├── Operator badge (AND/OR)
│   │   ├── Child conditions
│   │   │   ├── ConditionBuilder (simple condition)
│   │   │   ├── ConditionBuilder (simple condition)
│   │   │   └── ...
│   │   └── Action buttons
│   └── ...
└── Action buttons
```

## Recursive Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        CONDITIONBUILDER RENDERING FLOW                       │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ConditionBuilder({ form, path, ... })                                      │
│  ┌────────────────────────────┐                                             │
│  │ Get conditionType from form│                                             │
│  └────────────────┬───────────┘                                             │
│                   │                                                         │
│                   ▼                                                         │
│  ┌─────────────────────────────┐                                            │
│  │ conditionType === "simple"? │                                            │
│  └────────────┬────────────────┘                                            │
│               │                                                             │
│     ┌─────────┴─────────┐                                                   │
│     │                   │                                                   │
│     ▼                   ▼                                                   │
│  ┌────────────┐  ┌─────────────────────────┐                                │
│  │ Render     │  │ Render composite        │                                │
│  │ simple     │  │ condition UI            │                                │
│  │ condition  │  │                         │                                │
│  └────────────┘  └─────────────┬───────────┘                                │
│                                │                                            │
│                                ▼                                            │
│                  ┌─────────────────────────┐                                │
│                  │ Is expanded?            │                                │
│                  └─────────────┬───────────┘                                │
│                                │                                            │
│                                │  (if expanded)                             │
│                                ▼                                            │
│                  ┌─────────────────────────┐                                │
│                  │ Iterate child conditions│◄────┐                         │
│                  └─────────────┬───────────┘     │                         │
│                                │                 │                         │
│                                ▼                 │                         │
│                  ┌─────────────────────────┐     │                         │
│                  │ Recursive call:         │     │                         │
│                  │ <ConditionBuilder       │     │                         │
│                  │   path={`${path}.conditions.${index}`}                  │
│                  │   ...other props/>      │     │                         │
│                  └─────────────┬───────────┘     │                         │
│                                │                 │                         │
│                                │                 │                         │
│                                ▼                 │                         │
│                  ┌─────────────────────────┐     │                         │
│                  │ Next child condition    │─────┘                         │
│                  └─────────────────────────┘                                │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Data Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        CONDITION DATA MANAGEMENT                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  addSimpleCondition()                                                       │
│  ┌────────────────────────────┐                                             │
│  │ Get current condition      │                                             │
│  │ from form                  │                                             │
│  └────────────────┬───────────┘                                             │
│                   │                                                         │
│                   ▼                                                         │
│  ┌─────────────────────────────┐                                            │
│  │ Add new simple condition    │                                            │
│  │ to conditions array         │                                            │
│  └────────────────┬────────────┘                                            │
│                   │                                                         │
│                   ▼                                                         │
│  ┌─────────────────────────────┐                                            │
│  │ Update form with new        │                                            │
│  │ conditions array            │                                            │
│  └─────────────────────────────┘                                            │
│                                                                             │
│  addCompositeCondition()                                                    │
│  ┌────────────────────────────┐                                             │
│  │ Get current condition      │                                             │
│  │ from form                  │                                             │
│  └────────────────┬───────────┘                                             │
│                   │                                                         │
│                   ▼                                                         │
│  ┌─────────────────────────────┐                                            │
│  │ Create new composite with   │                                            │
│  │ opposite operator of parent │                                            │
│  └────────────────┬────────────┘                                            │
│                   │                                                         │
│                   ▼                                                         │
│  ┌─────────────────────────────┐                                            │
│  │ Add default simple condition│                                            │
│  │ as child of new composite   │                                            │
│  └────────────────┬────────────┘                                            │
│                   │                                                         │
│                   ▼                                                         │
│  ┌─────────────────────────────┐                                            │
│  │ Update form with new        │                                            │
│  │ conditions array            │                                            │
│  └─────────────────────────────┘                                            │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Key Implementation Details

### Path-Based Form Access

Each instance of the component uses a unique path string to access the corresponding part of the form data:

```tsx
// Root condition
<ConditionBuilder path="condition" />

// First child condition of root
<ConditionBuilder path="condition.conditions.0" />

// First child of first child
<ConditionBuilder path="condition.conditions.0.conditions.0" />
```

This path-based approach enables:
1. Each component to read/write its own part of the form
2. Form validation to work at any nesting level
3. React Hook Form to efficiently track all changes

### Recursive Child Rendering

The component uses this pattern to recursively render its child conditions:

```tsx
{conditions.map((_, index) => (
    <ConditionBuilder
        key={index}
        form={form}
        path={`${path}.conditions.${index}`}
        onRemove={() => removeCondition(index)}
        showConnector={index > 0}
        parentOperator={operator}
        isLast={index === conditions.length - 1}
    />
))}
```

### Visual Tree Representation

Visual elements that help represent the condition tree:
1. **Operator Badges**: Color-coded AND/OR badges (blue/orange)
2. **Connector Lines**: Vertical and horizontal lines showing relationships
3. **Expandable Groups**: Collapsible condition groups with toggle buttons
4. **Visual Hierarchy**: Nested conditions with connector indentation

## Mapping to Backend Condition Model

The condition structure created by the component maps exactly to the backend model:

```typescript
// Simple condition (leaf node)
{
    type: "simple",
    target: "source_ip",
    match_type: "equal",
    match_value: "192.168.1.1"
}

// Composite condition (container node)
{
    type: "composite",
    operator: "AND",
    conditions: [
        // Child conditions (simple or composite)
    ]
}
```

## Maintenance Guide

When modifying this component:

1. **Understand Recursion**: Changes may affect all nested levels
2. **Test Deep Nesting**: Validate changes with multi-level nested conditions
3. **Form Integration**: Maintain correct form paths for each condition
4. **Visual Elements**: Preserve visual connectors that show logical structure
5. **Performance**: Be aware of re-rendering issues in deeply nested structures

## Example Usage

```tsx
import { useForm } from "react-hook-form"
import { ConditionBuilder } from "./ConditionBuilder"
import type { MicroRuleCreateRequest } from "@/types/rule"

function RuleForm() {
    const form = useForm<MicroRuleCreateRequest>({
        defaultValues: {
            name: "",
            condition: {
                type: "composite",
                operator: "AND",
                conditions: [
                    {
                        type: "simple",
                        target: "source_ip",
                        match_type: "equal",
                        match_value: ""
                    }
                ]
            },
            // Other fields...
        }
    })
    
    return (
        <form>
            {/* Other form fields */}
            <ConditionBuilder 
                form={form} 
                path="condition" 
                isRoot={true} 
            />
            {/* Form submission */}
        </form>
    )
}
```

## Relationship with Backend Rule Engine

The frontend ConditionBuilder component and the backend MicroEngine implement the same condition model, forming a complete front-to-back rule system:

### Front-to-Back Mapping

| Frontend UI Element | Backend Entity | Description |
|---------------------|----------------|-------------|
| Simple condition form | SimpleCondition | Leaf node condition, directly matches target |
| Condition group | CompositeCondition | Container node, combines multiple conditions |
| AND/OR toggle | LogicalOperator | Determines logical relationship in condition group |
| Target selector | TargetType | Specifies type of target to match |
| Match type selector | MatchType | Defines how matching is performed |

### Recursion Pattern Comparison

Both frontend and backend use recursive patterns to handle the condition tree, but with different focuses:

1. **Frontend Recursion**:
   - Recursively renders UI component tree
   - Handles user interaction and visualization
   - Manages path-based form state

2. **Backend Recursion**:
   - Recursively parses condition structure
   - Recursively evaluates condition matches
   - Applies short-circuit logic for performance

### Data Flow

```
┌─────────────────┐    ┌────────────────┐    ┌─────────────────┐
│                 │    │                │    │                 │
│  Frontend       │───►│  API Request/  │───►│  Backend        │
│  ConditionBuilder    │  Response      │    │  MicroEngine    │
│                 │◄───│  (JSON data)   │◄───│                 │
└─────────────────┘    └────────────────┘    └─────────────────┘
```
