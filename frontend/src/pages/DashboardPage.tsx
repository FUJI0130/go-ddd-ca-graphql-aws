// src/pages/DashboardPage.tsx - 完全実装版

import React from 'react';
import { gql, useQuery } from '@apollo/client';
import { useAuth } from '../contexts/AuthContext';
import { MainLayout } from '../layouts/MainLayout';

// GraphQLスキーマ取得用のクエリ（簡易版）
const GET_SCHEMA_QUERY = gql`
  query GetSchema {
    __schema {
      queryType {
        name
      }
    }
  }
`;

export const DashboardPage: React.FC = () => {
    const { user, logout, isLoading: authLoading } = useAuth();
    const { data, loading, error } = useQuery(GET_SCHEMA_QUERY);

    const handleLogout = async () => {
        try {
            await logout();
        } catch (error) {
            console.error('ログアウトエラー:', error);
        }
    };

    if (authLoading) {
        return (
            <MainLayout>
                <div style={{ textAlign: 'center', padding: '20px' }}>
                    <p>🔄 認証状態を確認中...</p>
                </div>
            </MainLayout>
        );
    }

    return (
        <MainLayout>
            <div style={{ padding: '20px' }}>
                {/* ヘッダー */}
                <header style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    marginBottom: '30px',
                    padding: '0 0 20px 0',
                    borderBottom: '1px solid #ddd',
                }}>
                    <div>
                        <h1 style={{ margin: '0 0 8px 0' }}>テスト管理システム</h1>
                        <p style={{ margin: '0', color: '#666' }}>
                            ようこそ、{user?.username} さん ({user?.role})
                        </p>
                    </div>
                    <button
                        onClick={handleLogout}
                        style={{
                            padding: '8px 16px',
                            backgroundColor: '#dc3545',
                            color: 'white',
                            border: 'none',
                            borderRadius: '4px',
                            cursor: 'pointer',
                            fontSize: '14px',
                        }}
                    >
                        ログアウト
                    </button>
                </header>

                {/* 認証成功メッセージ */}
                <div style={{
                    padding: '16px',
                    marginBottom: '20px',
                    backgroundColor: '#d4edda',
                    border: '1px solid #c3e6cb',
                    borderRadius: '4px',
                    color: '#155724',
                }}>
                    <h3 style={{ margin: '0 0 8px 0' }}>✅ HttpOnly Cookie認証成功!</h3>
                    <p style={{ margin: '0' }}>
                        セキュアな認証システムが正常に動作しています。
                        ブラウザのDevTools → Application → Cookies で auth_token が確認できます。
                    </p>
                </div>

                {/* ユーザー情報カード */}
                <div style={{
                    padding: '20px',
                    border: '1px solid #ddd',
                    borderRadius: '8px',
                    marginBottom: '20px',
                    backgroundColor: '#f8f9fa',
                }}>
                    <h3 style={{ margin: '0 0 16px 0' }}>👤 ユーザー情報</h3>
                    <div style={{ display: 'grid', gridTemplateColumns: '120px 1fr', gap: '8px' }}>
                        <strong>ID:</strong> <span>{user?.id}</span>
                        <strong>ユーザー名:</strong> <span>{user?.username}</span>
                        <strong>ロール:</strong> <span>{user?.role}</span>
                        <strong>作成日:</strong> <span>{user?.createdAt ? new Date(user.createdAt).toLocaleString() : '-'}</span>
                        <strong>最終ログイン:</strong> <span>{user?.lastLoginAt ? new Date(user.lastLoginAt).toLocaleString() : '-'}</span>
                    </div>
                </div>

                {/* GraphQL接続テスト */}
                <div style={{
                    padding: '20px',
                    border: '1px solid #ddd',
                    borderRadius: '8px',
                    marginBottom: '20px',
                }}>
                    <h3 style={{ margin: '0 0 16px 0' }}>🔗 GraphQL接続テスト</h3>
                    {loading && <p>🔄 GraphQLサーバーに接続中...</p>}
                    {error && <p style={{ color: '#dc3545' }}>❌ エラー: {error.message}</p>}
                    {data && (
                        <p style={{ color: '#28a745' }}>
                            ✅ GraphQL接続成功! Query Type: {data.__schema.queryType.name}
                        </p>
                    )}
                </div>

                {/* 実装状況 */}
                <div style={{
                    padding: '20px',
                    border: '1px solid #ddd',
                    borderRadius: '8px',
                    backgroundColor: '#e3f2fd',
                }}>
                    <h3 style={{ margin: '0 0 16px 0' }}>🚀 実装完了機能</h3>
                    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: '12px' }}>
                        <div>
                            <h4 style={{ margin: '0 0 8px 0', color: '#1976d2' }}>✅ 認証システム</h4>
                            <ul style={{ margin: '0', paddingLeft: '20px' }}>
                                <li>HttpOnly Cookie認証</li>
                                <li>JWT トークン管理</li>
                                <li>認証状態管理</li>
                                <li>ログイン・ログアウト</li>
                            </ul>
                        </div>
                        <div>
                            <h4 style={{ margin: '0 0 8px 0', color: '#1976d2' }}>✅ フロントエンド基盤</h4>
                            <ul style={{ margin: '0', paddingLeft: '20px' }}>
                                <li>React + TypeScript</li>
                                <li>Apollo Client統合</li>
                                <li>GraphQL Code Generator</li>
                                <li>認証コンテキスト</li>
                            </ul>
                        </div>
                        <div>
                            <h4 style={{ margin: '0 0 8px 0', color: '#1976d2' }}>✅ React Router統合</h4>
                            <ul style={{ margin: '0', paddingLeft: '20px' }}>
                                <li>URL-based routing</li>
                                <li>認証保護ルート</li>
                                <li>Page-based アーキテクチャ</li>
                                <li>プロフェッショナル設計</li>
                            </ul>
                        </div>
                    </div>
                </div>
            </div>
        </MainLayout>
    );
};