# ConditionBuilder 组件文档

## 概述 (Overview)

`ConditionBuilder` 组件是一个递归式 React 组件，用于在 WAF 界面中创建和可视化复杂的嵌套条件层次结构。它提供了一个直观的 UI，用于构建简单条件和具有任意嵌套深度的复合条件。

## 组件架构 (Component Architecture)

该组件实现了自递归模式来渲染条件树，与后端的规则引擎条件模型相匹配。

### 核心特性 (Core Features)

1. **递归渲染 (Recursive Rendering)**: 可以将自身作为子组件渲染，创建嵌套的条件组
2. **两种条件类型 (Two Condition Types)**: 支持简单条件（叶节点）和复合条件（容器节点）
3. **动态条件管理 (Dynamic Condition Management)**: 允许通过 UI 添加、删除和配置条件
4. **逻辑运算符 (Logical Operators)**: 支持 AND/OR 运算符，并提供视觉指示器
5. **视觉连接 (Visual Connectivity)**: 显示连接线，直观表示逻辑层次结构

## 递归 UI 结构 (Recursive UI Structure)

```
ConditionBuilder (根 - 复合条件)
├── 运算符徽章 (AND/OR)
├── 子条件 (数组)
│   ├── ConditionBuilder (简单条件)
│   ├── ConditionBuilder (复合条件)
│   │   ├── 运算符徽章 (AND/OR)
│   │   ├── 子条件
│   │   │   ├── ConditionBuilder (简单条件)
│   │   │   ├── ConditionBuilder (简单条件)
│   │   │   └── ...
│   │   └── 操作按钮
│   └── ...
└── 操作按钮
```

## 递归流程图 (Recursive Flow Diagram)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        CONDITIONBUILDER 渲染流程                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ConditionBuilder({ form, path, ... })                                      │
│  ┌────────────────────────────┐                                             │
│  │ 从表单获取 conditionType   │                                             │
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
│  │ 渲染简单   │  │ 渲染复合条件 UI         │                                │
│  │ 条件 UI    │  │                         │                                │
│  └────────────┘  └─────────────┬───────────┘                                │
│                                │                                            │
│                                ▼                                            │
│                  ┌─────────────────────────┐                                │
│                  │ 是否展开？              │                                │
│                  └─────────────┬───────────┘                                │
│                                │                                            │
│                                │  (如果展开)                                │
│                                ▼                                            │
│                  ┌─────────────────────────┐                                │
│                  │ 遍历子条件              │◄────┐                         │
│                  └─────────────┬───────────┘     │                         │
│                                │                 │                         │
│                                ▼                 │                         │
│                  ┌─────────────────────────┐     │                         │
│                  │ 递归调用:               │     │                         │
│                  │ <ConditionBuilder       │     │                         │
│                  │   path={`${path}.conditions.${index}`}                  │
│                  │   ...其他属性/>         │     │                         │
│                  └─────────────┬───────────┘     │                         │
│                                │                 │                         │
│                                │                 │                         │
│                                ▼                 │                         │
│                  ┌─────────────────────────┐     │                         │
│                  │ 下一个子条件            │─────┘                         │
│                  └─────────────────────────┘                                │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## 数据流 (Data Flow)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        条件数据管理                                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  addSimpleCondition()                                                       │
│  ┌────────────────────────────┐                                             │
│  │ 从表单获取当前条件         │                                             │
│  │                            │                                             │
│  └────────────────┬───────────┘                                             │
│                   │                                                         │
│                   ▼                                                         │
│  ┌─────────────────────────────┐                                            │
│  │ 将新的简单条件添加到        │                                            │
│  │ 条件数组                    │                                            │
│  └────────────────┬────────────┘                                            │
│                   │                                                         │
│                   ▼                                                         │
│  ┌─────────────────────────────┐                                            │
│  │ 用新的条件数组更新表单      │                                            │
│  │                             │                                            │
│  └─────────────────────────────┘                                            │
│                                                                             │
│  addCompositeCondition()                                                    │
│  ┌────────────────────────────┐                                             │
│  │ 从表单获取当前条件         │                                             │
│  │                            │                                             │
│  └────────────────┬───────────┘                                             │
│                   │                                                         │
│                   ▼                                                         │
│  ┌─────────────────────────────┐                                            │
│  │ 创建使用与父级相反运算符    │                                            │
│  │ 的新复合条件                │                                            │
│  └────────────────┬────────────┘                                            │
│                   │                                                         │
│                   ▼                                                         │
│  ┌─────────────────────────────┐                                            │
│  │ 添加默认简单条件作为        │                                            │
│  │ 新复合条件的子条件          │                                            │
│  └────────────────┬────────────┘                                            │
│                   │                                                         │
│                   ▼                                                         │
│  ┌─────────────────────────────┐                                            │
│  │ 用新的条件数组更新表单      │                                            │
│  │                             │                                            │
│  └─────────────────────────────┘                                            │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## 关键实现细节 (Key Implementation Details)

### 基于路径的表单访问 (Path-Based Form Access)

组件的每个实例使用唯一的路径字符串来访问表单数据的相应部分：

```tsx
// 根条件
<ConditionBuilder path="condition" />

// 根条件的第一个子条件
<ConditionBuilder path="condition.conditions.0" />

// 第一个子条件的第一个子条件
<ConditionBuilder path="condition.conditions.0.conditions.0" />
```

这种基于路径的方法使得：
1. 每个组件都可以读取/写入表单的自己那部分
2. 表单验证可在任何嵌套级别正常工作
3. React Hook Form 可以高效跟踪所有变更

### 递归子条件渲染 (Recursive Child Rendering)

组件使用这种模式递归渲染其子条件：

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

### 树形结构可视化 (Visual Tree Representation)

帮助表示条件树的视觉元素：
1. **运算符徽章 (Operator Badges)**: 不同颜色的 AND/OR 徽章（蓝色/橙色）
2. **连接线 (Connector Lines)**: 显示关系的垂直和水平线
3. **可展开组 (Expandable Groups)**: 带有切换按钮的可折叠条件组
4. **视觉层次 (Visual Hierarchy)**: 嵌套条件带有连接器缩进

## 与后端条件模型的映射 (Mapping to Backend Condition Model)

组件创建的条件结构与后端模型完全匹配：

```typescript
// 简单条件 (叶节点)
{
    type: "simple",
    target: "source_ip",
    match_type: "equal",
    match_value: "192.168.1.1"
}

// 复合条件 (容器节点)
{
    type: "composite",
    operator: "AND",
    conditions: [
        // 子条件 (简单或复合)
    ]
}
```

## 维护指南 (Maintenance Guide)

修改此组件时：

1. **理解递归**: 更改可能影响所有嵌套级别
2. **测试深层嵌套**: 使用多级嵌套条件验证更改
3. **表单集成**: 为每个条件维护正确的表单路径
4. **视觉元素**: 保留显示逻辑结构的视觉连接器
5. **性能**: 注意深度嵌套结构中的重新渲染问题

## 使用示例 (Example Usage)

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
            // 其他字段...
        }
    })
    
    return (
        <form>
            {/* 其他表单字段 */}
            <ConditionBuilder 
                form={form} 
                path="condition" 
                isRoot={true} 
            />
            {/* 表单提交 */}
        </form>
    )
}
```

## ConditionBuilder 与后端规则引擎的关系 (Relationship with Backend Rule Engine)

前端 ConditionBuilder 组件与后端微引擎（MicroEngine）实现了相同的条件模型，形成一个完整的前后端规则系统：

### 前后端条件映射关系 (Front-to-Back Mapping)

| 前端 UI 元素 | 后端对应实体 | 说明 |
|------------|------------|------|
| 简单条件表单 | SimpleCondition | 叶节点条件，直接匹配目标 |
| 条件组 | CompositeCondition | 容器节点，组合多个条件 |
| AND/OR 切换 | LogicalOperator | 决定条件组的逻辑关系 |
| 目标选择器 | TargetType | 指定匹配的目标类型 |
| 匹配类型选择器 | MatchType | 定义匹配的方式 |

### 递归模式对比 (Recursion Pattern Comparison)

前端和后端都使用递归模式处理条件树，但各自有不同的侧重点：

1. **前端递归**:
   - 递归渲染UI组件树
   - 处理用户交互和可视化
   - 管理基于路径的表单状态

2. **后端递归**:
   - 递归解析条件结构
   - 递归评估条件匹配
   - 应用短路逻辑提升性能

### 数据流向 (Data Flow)

```
┌─────────────────┐    ┌────────────────┐    ┌─────────────────┐
│                 │    │                │    │                 │
│  前端条件构建器  │───►│  API 请求/响应  │───►│  后端规则引擎   │
│  (ConditionBuilder)  │  (JSON 数据)   │    │  (MicroEngine)  │
│                 │◄───│                │◄───│                 │
└─────────────────┘    └────────────────┘    └─────────────────┘
```
