// src/pages/TestSuiteListPage.tsx - 完全実装版

import React, { useState } from 'react';
import { useTestSuites, useUpdateTestSuiteStatus } from '../hooks/useTestSuites';
import { TestSuiteList } from '../components/testSuite/TestSuiteList';
import { TestSuiteFilters as FilterModal } from '../components/testSuite/TestSuiteFilters';
import { CreateTestSuiteModal } from '../components/testSuite/CreateTestSuiteModal';
import { MainLayout } from '../layouts/MainLayout';
import { TestSuite, TestSuiteFilters } from '../types/testSuite';
// Generated SuiteStatusを使用
import { SuiteStatus } from '../generated/graphql';

export const TestSuiteListPage: React.FC = () => {
    // カスタムフックによるデータ管理
    const {
        testSuites,
        totalCount,
        loading,
        error,
        filters,
        hasNextPage,
        loadMore,
        updateFilters,
        refresh
    } = useTestSuites({ pageSize: 10 });

    const { updateStatus, loading: statusLoading } = useUpdateTestSuiteStatus();

    // UI状態管理（ローカル状態）
    const [showCreateModal, setShowCreateModal] = useState(false);
    const [showFilters, setShowFilters] = useState(false);

    // イベントハンドラー
    const handleStatusChange = async (id: string, status: SuiteStatus) => {
        try {
            await updateStatus(id, status);
            // 成功通知（実際にはNotificationContextを使用）
            console.log('ステータス更新成功');
        } catch (error) {
            console.error('ステータス更新失敗:', error);
        }
    };

    const handleFilterChange = (newFilters: Partial<TestSuiteFilters>) => {
        updateFilters(newFilters);
    };

    const handleCreateSuccess = () => {
        setShowCreateModal(false);
        refresh(); // 一覧を更新
    };

    const handleSelectSuite = (suite: TestSuite) => {
        // 詳細画面への遷移（実際にはルーティングを使用）
        console.log('テストスイート選択:', suite.id);
    };

    // refresh関数をラップしてMouseEvent対応
    const handleRefresh = () => {
        refresh();
    };

    // エラー状態の表示
    if (error) {
        return (
            <MainLayout>
                <div style={{
                    padding: '20px',
                    textAlign: 'center',
                    backgroundColor: '#ffebee',
                    border: '1px solid #ffcdd2',
                    borderRadius: '4px',
                    margin: '20px'
                }}>
                    <h3 style={{ color: '#c62828', margin: '0 0 16px 0' }}>
                        ❌ データの読み込みに失敗しました
                    </h3>
                    <p style={{ margin: '0 0 16px 0', color: '#666' }}>
                        {error.message || 'ネットワークエラーが発生しました'}
                    </p>
                    <button
                        onClick={handleRefresh}
                        style={{
                            padding: '8px 16px',
                            backgroundColor: '#007bff',
                            color: 'white',
                            border: 'none',
                            borderRadius: '4px',
                            cursor: 'pointer'
                        }}
                    >
                        再試行
                    </button>
                </div>
            </MainLayout>
        );
    }

    return (
        <MainLayout>
            <div style={{ padding: '20px' }}>
                {/* ヘッダー部分 */}
                <div style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    marginBottom: '20px',
                    paddingBottom: '16px',
                    borderBottom: '1px solid #ddd'
                }}>
                    <div>
                        <h2 style={{ margin: '0 0 8px 0' }}>テストスイート管理</h2>
                        <p style={{ margin: '0', color: '#666' }}>
                            総件数: {totalCount}件
                            {filters.status && ` (${filters.status}でフィルター中)`}
                            {filters.search && ` (「${filters.search}」で検索中)`}
                        </p>
                    </div>

                    <div style={{ display: 'flex', gap: '12px' }}>
                        <button
                            onClick={() => setShowFilters(true)}
                            style={{
                                padding: '8px 16px',
                                backgroundColor: '#6c757d',
                                color: 'white',
                                border: 'none',
                                borderRadius: '4px',
                                cursor: 'pointer'
                            }}
                        >
                            🔍 フィルター
                        </button>

                        <button
                            onClick={() => setShowCreateModal(true)}
                            style={{
                                padding: '8px 16px',
                                backgroundColor: '#28a745',
                                color: 'white',
                                border: 'none',
                                borderRadius: '4px',
                                cursor: 'pointer'
                            }}
                        >
                            ➕ 新規作成
                        </button>
                    </div>
                </div>

                {/* 一覧コンポーネント（Layer 4に委譲） */}
                <TestSuiteList
                    testSuites={testSuites}
                    loading={loading}
                    statusLoading={statusLoading}
                    hasNextPage={hasNextPage}
                    onStatusChange={handleStatusChange}
                    onSelectSuite={handleSelectSuite}
                    onLoadMore={loadMore}
                    onRefresh={handleRefresh}
                />

                {/* モーダル群 */}
                {showCreateModal && (
                    <CreateTestSuiteModal
                        onClose={() => setShowCreateModal(false)}
                        onSuccess={handleCreateSuccess}
                    />
                )}

                {showFilters && (
                    <FilterModal
                        currentFilters={filters}
                        onApply={handleFilterChange}
                        onClose={() => setShowFilters(false)}
                    />
                )}
            </div>
        </MainLayout>
    );
};