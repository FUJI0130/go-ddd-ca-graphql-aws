// src/components/testSuite/TestSuiteList.tsx

import React from 'react';
import { TestSuite } from '../../types/testSuite';
// Generated SuiteStatusを使用
import { SuiteStatus } from '../../generated/graphql';

interface TestSuiteListProps {
    testSuites: TestSuite[];
    loading: boolean;
    statusLoading: boolean;
    hasNextPage: boolean;
    onStatusChange: (id: string, status: SuiteStatus) => void;
    onSelectSuite: (suite: TestSuite) => void;
    onLoadMore: () => void;
    onRefresh: () => void;
}

// ステータス表示用ヘルパー（Generated SuiteStatus対応）
const getStatusInfo = (status: SuiteStatus) => {
    const statusMap = {
        [SuiteStatus.Preparation]: { label: '準備中', color: '#6c757d', emoji: '📋' },
        [SuiteStatus.InProgress]: { label: '実行中', color: '#007bff', emoji: '🔄' },
        [SuiteStatus.Completed]: { label: '完了', color: '#28a745', emoji: '✅' },
        [SuiteStatus.Suspended]: { label: '一時停止', color: '#ffc107', emoji: '⏸️' }
    };
    return statusMap[status] || { label: status, color: '#6c757d', emoji: '❓' };
};

// 進捗バー表示コンポーネント
const ProgressBar: React.FC<{ progress: number }> = ({ progress }) => (
    <div style={{
        width: '100%',
        height: '8px',
        backgroundColor: '#e9ecef',
        borderRadius: '4px',
        overflow: 'hidden'
    }}>
        <div style={{
            width: `${Math.min(Math.max(progress, 0), 100)}%`,
            height: '100%',
            backgroundColor: progress >= 100 ? '#28a745' : '#007bff',
            borderRadius: '4px',
            transition: 'width 0.3s ease'
        }} />
    </div>
);

// 個別テストスイートカードコンポーネント
const TestSuiteCard: React.FC<{
    suite: TestSuite;
    statusLoading: boolean;
    onStatusChange: (id: string, status: SuiteStatus) => void;
    onSelect: (suite: TestSuite) => void;
}> = ({ suite, statusLoading, onStatusChange, onSelect }) => {
    const statusInfo = getStatusInfo(suite.status);

    return (
        <div style={{
            border: '1px solid #ddd',
            borderRadius: '8px',
            padding: '16px',
            marginBottom: '12px',
            backgroundColor: 'white',
            boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
            transition: 'box-shadow 0.2s',
            cursor: 'pointer'
        }}
            onMouseEnter={(e) => {
                e.currentTarget.style.boxShadow = '0 4px 8px rgba(0,0,0,0.15)';
            }}
            onMouseLeave={(e) => {
                e.currentTarget.style.boxShadow = '0 2px 4px rgba(0,0,0,0.1)';
            }}
            onClick={() => onSelect(suite)}
        >
            <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'flex-start',
                marginBottom: '12px'
            }}>
                <div style={{ flex: 1 }}>
                    <h3 style={{ margin: '0 0 8px 0', color: '#333' }}>
                        {suite.name}
                    </h3>
                    <p style={{ margin: '0 0 8px 0', color: '#666', fontSize: '14px' }}>
                        {suite.description || '説明なし'}
                    </p>
                </div>

                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                    <span style={{
                        padding: '4px 8px',
                        borderRadius: '12px',
                        fontSize: '12px',
                        fontWeight: 'bold',
                        color: 'white',
                        backgroundColor: statusInfo.color
                    }}>
                        {statusInfo.emoji} {statusInfo.label}
                    </span>

                    <select
                        value={suite.status}
                        onChange={(e) => {
                            e.stopPropagation();
                            onStatusChange(suite.id, e.target.value as SuiteStatus);
                        }}
                        disabled={statusLoading}
                        style={{
                            padding: '4px 8px',
                            border: '1px solid #ddd',
                            borderRadius: '4px',
                            fontSize: '12px',
                            cursor: statusLoading ? 'not-allowed' : 'pointer'
                        }}
                    >
                        <option value={SuiteStatus.Preparation}>準備中</option>
                        <option value={SuiteStatus.InProgress}>実行中</option>
                        <option value={SuiteStatus.Completed}>完了</option>
                        <option value={SuiteStatus.Suspended}>一時停止</option>
                    </select>
                </div>
            </div>

            {/* 進捗情報 */}
            <div style={{ marginBottom: '12px' }}>
                <div style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    marginBottom: '4px'
                }}>
                    <span style={{ fontSize: '12px', color: '#666' }}>進捗</span>
                    <span style={{ fontSize: '12px', fontWeight: 'bold' }}>
                        {suite.progress}%
                    </span>
                </div>
                <ProgressBar progress={suite.progress} />
            </div>

            {/* 日付情報 */}
            <div style={{
                display: 'grid',
                gridTemplateColumns: '1fr 1fr',
                gap: '8px',
                fontSize: '12px',
                color: '#666'
            }}>
                <div>
                    <strong>開始予定:</strong><br />
                    {new Date(suite.estimatedStartDate).toLocaleDateString()}
                </div>
                <div>
                    <strong>終了予定:</strong><br />
                    {new Date(suite.estimatedEndDate).toLocaleDateString()}
                </div>
            </div>
        </div>
    );
};

// メインの一覧コンポーネント（Layer 4: プレゼンテーション）
export const TestSuiteList: React.FC<TestSuiteListProps> = ({
    testSuites,
    loading,
    statusLoading,
    hasNextPage,
    onStatusChange,
    onSelectSuite,
    onLoadMore,
    onRefresh
}) => {
    // ローディング状態
    if (loading && testSuites.length === 0) {
        return (
            <div style={{ textAlign: 'center', padding: '40px' }}>
                <div style={{ fontSize: '24px', marginBottom: '16px' }}>🔄</div>
                <p>テストスイートを読み込み中...</p>
            </div>
        );
    }

    // 空状態
    if (!loading && testSuites.length === 0) {
        return (
            <div style={{
                textAlign: 'center',
                padding: '40px',
                backgroundColor: '#f8f9fa',
                border: '2px dashed #dee2e6',
                borderRadius: '8px'
            }}>
                <div style={{ fontSize: '48px', marginBottom: '16px' }}>📝</div>
                <h3 style={{ margin: '0 0 16px 0', color: '#666' }}>
                    テストスイートがありません
                </h3>
                <p style={{ margin: '0', color: '#666' }}>
                    新しいテストスイートを作成してください
                </p>
            </div>
        );
    }

    return (
        <div>
            {/* テストスイート一覧 */}
            {testSuites.map((suite) => (
                <TestSuiteCard
                    key={suite.id}
                    suite={suite}
                    statusLoading={statusLoading}
                    onStatusChange={onStatusChange}
                    onSelect={onSelectSuite}
                />
            ))}

            {/* 追加読み込みボタン */}
            {hasNextPage && (
                <div style={{ textAlign: 'center', marginTop: '20px' }}>
                    <button
                        onClick={onLoadMore}
                        disabled={loading}
                        style={{
                            padding: '12px 24px',
                            backgroundColor: loading ? '#ccc' : '#007bff',
                            color: 'white',
                            border: 'none',
                            borderRadius: '4px',
                            cursor: loading ? 'not-allowed' : 'pointer',
                            fontSize: '14px'
                        }}
                    >
                        {loading ? '🔄 読み込み中...' : '📄 さらに読み込む'}
                    </button>
                </div>
            )}

            {/* 読み込み中インジケーター（追加読み込み時） */}
            {loading && testSuites.length > 0 && (
                <div style={{
                    textAlign: 'center',
                    padding: '20px',
                    color: '#666'
                }}>
                    <div style={{ fontSize: '20px', marginBottom: '8px' }}>🔄</div>
                    <p style={{ margin: '0' }}>追加データを読み込み中...</p>
                </div>
            )}

            {/* リフレッシュボタン */}
            <div style={{
                textAlign: 'center',
                marginTop: '20px',
                paddingTop: '20px',
                borderTop: '1px solid #eee'
            }}>
                <button
                    onClick={onRefresh}
                    disabled={loading}
                    style={{
                        padding: '8px 16px',
                        backgroundColor: '#6c757d',
                        color: 'white',
                        border: 'none',
                        borderRadius: '4px',
                        cursor: loading ? 'not-allowed' : 'pointer',
                        fontSize: '12px'
                    }}
                >
                    🔄 リフレッシュ
                </button>
            </div>
        </div>
    );
};