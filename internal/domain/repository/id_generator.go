package repository

// このパッケージのIDジェネレーターインターフェースは、
// ドメインモデルの階層構造を反映した意図的に異なるシグネチャを持っています。
// これはエンティティ間の親子関係を表現するために設計されています。

// TestSuiteIDGenerator はテストスイートIDを生成するインターフェース
// テストスイートは最上位エンティティなので、他のIDに依存しません。
type TestSuiteIDGenerator interface {
	GenerateID() (string, error)
}

// TestGroupIDGenerator はテストグループIDを生成するインターフェース
// テストグループはテストスイートに属するため、スイートIDを必要とします。
type TestGroupIDGenerator interface {
	GenerateID(suiteID string) (string, error)
}

// TestCaseIDGenerator はテストケースIDを生成するインターフェース
// テストケースはテストグループに属するため、グループIDを必要とします。
type TestCaseIDGenerator interface {
	GenerateID(groupID string) (string, error)
}
