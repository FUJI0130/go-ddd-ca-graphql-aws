// src/components/testSuite/TestSuiteFilters.tsx

import React, { useState } from 'react';
import { TestSuiteFilters as FilterType } from '../../types/testSuite';
// Generated SuiteStatusを直接import
import { SuiteStatus } from '../../generated/graphql';

interface TestSuiteFiltersProps {
    currentFilters: FilterType;
    onApply: (filters: Partial<FilterType>) => void;
    onClose: () => void;
}

export const TestSuiteFilters: React.FC<TestSuiteFiltersProps> = ({
    currentFilters,
    onApply,
    onClose
}) => {
    const [localFilters, setLocalFilters] = useState<FilterType>(currentFilters);

    const handleApply = () => {
        onApply(localFilters);
        onClose();
    };

    const handleReset = () => {
        const resetFilters: FilterType = {};
        setLocalFilters(resetFilters);
        onApply(resetFilters);
        onClose();
    };

    return (
        <div style={{
            position: 'fixed',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            backgroundColor: 'rgba(0,0,0,0.5)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            zIndex: 1000
        }}>
            <div style={{
                backgroundColor: 'white',
                padding: '24px',
                borderRadius: '8px',
                width: '400px',
                maxWidth: '90vw'
            }}>
                <h3 style={{ margin: '0 0 20px 0' }}>フィルター設定</h3>

                {/* ステータスフィルター */}
                <div style={{ marginBottom: '16px' }}>
                    <label style={{ display: 'block', marginBottom: '8px', fontWeight: 'bold' }}>
                        ステータス
                    </label>
                    <select
                        value={localFilters.status || ''}
                        onChange={(e) =>
                            setLocalFilters(prev => ({
                                ...prev,
                                status: e.target.value as SuiteStatus || undefined
                            }))
                        }
                        style={{
                            width: '100%',
                            padding: '8px',
                            border: '1px solid #ddd',
                            borderRadius: '4px'
                        }}
                    >
                        <option value="">すべて</option>
                        <option value={SuiteStatus.Preparation}>準備中</option>
                        <option value={SuiteStatus.InProgress}>実行中</option>
                        <option value={SuiteStatus.Completed}>完了</option>
                        <option value={SuiteStatus.Suspended}>一時停止</option>
                    </select>
                </div>

                {/* 検索キーワード */}
                <div style={{ marginBottom: '16px' }}>
                    <label style={{ display: 'block', marginBottom: '8px', fontWeight: 'bold' }}>
                        検索キーワード
                    </label>
                    <input
                        type="text"
                        value={localFilters.search || ''}
                        onChange={(e) =>
                            setLocalFilters(prev => ({
                                ...prev,
                                search: e.target.value || undefined
                            }))
                        }
                        placeholder="名前や説明で検索"
                        style={{
                            width: '100%',
                            padding: '8px',
                            border: '1px solid #ddd',
                            borderRadius: '4px'
                        }}
                    />
                </div>

                {/* 期間フィルター */}
                <div style={{ marginBottom: '20px' }}>
                    <label style={{ display: 'block', marginBottom: '8px', fontWeight: 'bold' }}>
                        期間
                    </label>
                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '8px' }}>
                        <div>
                            <label style={{ fontSize: '12px', color: '#666' }}>開始日から</label>
                            <input
                                type="date"
                                value={localFilters.dateFrom || ''}
                                onChange={(e) =>
                                    setLocalFilters(prev => ({
                                        ...prev,
                                        dateFrom: e.target.value || undefined
                                    }))
                                }
                                style={{
                                    width: '100%',
                                    padding: '8px',
                                    border: '1px solid #ddd',
                                    borderRadius: '4px'
                                }}
                            />
                        </div>
                        <div>
                            <label style={{ fontSize: '12px', color: '#666' }}>終了日まで</label>
                            <input
                                type="date"
                                value={localFilters.dateTo || ''}
                                onChange={(e) =>
                                    setLocalFilters(prev => ({
                                        ...prev,
                                        dateTo: e.target.value || undefined
                                    }))
                                }
                                style={{
                                    width: '100%',
                                    padding: '8px',
                                    border: '1px solid #ddd',
                                    borderRadius: '4px'
                                }}
                            />
                        </div>
                    </div>
                </div>

                {/* ボタン群 */}
                <div style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    gap: '8px'
                }}>
                    <button
                        onClick={handleReset}
                        style={{
                            padding: '8px 16px',
                            backgroundColor: '#6c757d',
                            color: 'white',
                            border: 'none',
                            borderRadius: '4px',
                            cursor: 'pointer'
                        }}
                    >
                        リセット
                    </button>

                    <div style={{ display: 'flex', gap: '8px' }}>
                        <button
                            onClick={onClose}
                            style={{
                                padding: '8px 16px',
                                backgroundColor: 'transparent',
                                color: '#666',
                                border: '1px solid #ddd',
                                borderRadius: '4px',
                                cursor: 'pointer'
                            }}
                        >
                            キャンセル
                        </button>

                        <button
                            onClick={handleApply}
                            style={{
                                padding: '8px 16px',
                                backgroundColor: '#007bff',
                                color: 'white',
                                border: 'none',
                                borderRadius: '4px',
                                cursor: 'pointer'
                            }}
                        >
                            適用
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
};