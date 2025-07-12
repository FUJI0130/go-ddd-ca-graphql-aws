# コンテキストマップ

## システム概要
テストケース管理システムの境界と関連性を示すコンテキストマップです。

## 図解
```mermaid
graph LR
    subgraph Core Domain
        PM[Project Management]
        TM[Task Management]
        GT[Gantt Timeline]
    end
    
    subgraph Supporting Domains
        TS[Test Suite Management]
        UM[User Management]
        NT[Notification]
    end
    
    subgraph Generic Domains
        GI[Git Integration]
        CI[CI Pipeline Integration]
        AN[Analytics]
    end
    
    PM --> TM
    PM --> GT
    TM --> GT
    TM --> TS
    PM --> UM
    TM --> NT
    TS --> NT
    GT --> GI
    TS --> CI
    PM --> AN