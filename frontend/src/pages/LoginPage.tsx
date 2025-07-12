// src/pages/LoginPage.tsx - 統合版（Page-based アーキテクチャ100%適用）

import React, { useState } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { LoginCredentials } from '../types/auth';

interface LocationState {
    from?: Location;
}

export const LoginPage: React.FC = () => {
    // 【維持】認証状態確認機能（既存LoginPage.tsxから）
    const { isAuthenticated, isLoading } = useAuth();
    const location = useLocation();
    const state = location.state as LocationState;

    // 【統合】ログイン処理機能（LoginForm.tsxから移行）
    const { login, error, resetAuthError } = useAuth();
    const [credentials, setCredentials] = useState<LoginCredentials>({
        username: '',
        password: '',
    });
    const [validationErrors, setValidationErrors] = useState<{
        username?: string;
        password?: string;
    }>({});

    // 【統合】フォーム処理ロジック（LoginForm.tsxから移行）

    // 入力値変更ハンドラー
    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const { name, value } = e.target;
        setCredentials(prev => ({
            ...prev,
            [name]: value,
        }));

        // バリデーションエラーをクリア
        if (validationErrors[name as keyof typeof validationErrors]) {
            setValidationErrors(prev => ({
                ...prev,
                [name]: undefined,
            }));
        }

        // 認証エラーをクリア
        if (error) {
            resetAuthError();
        }
    };

    // バリデーション
    const validateForm = (): boolean => {
        const errors: typeof validationErrors = {};

        if (!credentials.username.trim()) {
            errors.username = 'ユーザー名を入力してください';
        }

        if (!credentials.password.trim()) {
            errors.password = 'パスワードを入力してください';
        } else if (credentials.password.length < 6) {
            errors.password = 'パスワードは6文字以上で入力してください';
        }

        setValidationErrors(errors);
        return Object.keys(errors).length === 0;
    };

    // ログイン送信ハンドラー
    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!validateForm()) {
            return;
        }

        try {
            await login(credentials);
            // ログイン成功時はコンテキストが状態を管理するため、特別な処理は不要
        } catch (error) {
            // エラーはコンテキストで管理されるため、特別な処理は不要
        }
    };

    // キーボードショートカット（Enter）
    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter' && !isLoading) {
            handleSubmit(e as any);
        }
    };

    // 【維持】認証状態による分岐処理（既存LoginPage.tsxから）

    // 認証状態確認中
    if (isLoading) {
        return (
            <div style={{
                minHeight: '100vh',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                backgroundColor: '#f5f5f5',
                flexDirection: 'column',
                gap: '16px'
            }}>
                <div style={{ fontSize: '24px' }}>🔄</div>
                <p>認証状態を確認中...</p>
            </div>
        );
    }

    // 認証済みの場合は元のページまたはダッシュボードにリダイレクト
    if (isAuthenticated) {
        const from = state?.from?.pathname || '/';
        return <Navigate to={from} replace />;
    }

    // 【統合】完全なログインフォーム実装（LoginForm.tsxから移行）
    return (
        <div style={{
            minHeight: '100vh',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            backgroundColor: '#f5f5f5',
            padding: '20px'
        }}>
            {/* ログインフォーム完全実装 */}
            <div style={{
                maxWidth: '400px',
                margin: '0 auto',
                padding: '20px',
                border: '1px solid #ddd',
                borderRadius: '8px',
                backgroundColor: '#f9f9f9'
            }}>
                <h2 style={{ textAlign: 'center', marginBottom: '24px' }}>
                    ログイン
                </h2>

                <form onSubmit={handleSubmit}>
                    {/* ユーザー名入力 */}
                    <div style={{ marginBottom: '16px' }}>
                        <label htmlFor="username" style={{
                            display: 'block',
                            marginBottom: '4px',
                            fontWeight: 'bold'
                        }}>
                            ユーザー名
                        </label>
                        <input
                            type="text"
                            id="username"
                            name="username"
                            value={credentials.username}
                            onChange={handleInputChange}
                            onKeyDown={handleKeyDown}
                            disabled={isLoading}
                            style={{
                                width: '100%',
                                padding: '8px 12px',
                                border: `1px solid ${validationErrors.username ? '#ff6b6b' : '#ddd'}`,
                                borderRadius: '4px',
                                fontSize: '14px',
                                backgroundColor: isLoading ? '#f5f5f5' : 'white',
                            }}
                            placeholder="ユーザー名を入力"
                            autoComplete="username"
                            autoFocus
                        />
                        {validationErrors.username && (
                            <p style={{ color: '#ff6b6b', fontSize: '12px', margin: '4px 0 0 0' }}>
                                {validationErrors.username}
                            </p>
                        )}
                    </div>

                    {/* パスワード入力 */}
                    <div style={{ marginBottom: '16px' }}>
                        <label htmlFor="password" style={{
                            display: 'block',
                            marginBottom: '4px',
                            fontWeight: 'bold'
                        }}>
                            パスワード
                        </label>
                        <input
                            type="password"
                            id="password"
                            name="password"
                            value={credentials.password}
                            onChange={handleInputChange}
                            onKeyDown={handleKeyDown}
                            disabled={isLoading}
                            style={{
                                width: '100%',
                                padding: '8px 12px',
                                border: `1px solid ${validationErrors.password ? '#ff6b6b' : '#ddd'}`,
                                borderRadius: '4px',
                                fontSize: '14px',
                                backgroundColor: isLoading ? '#f5f5f5' : 'white',
                            }}
                            placeholder="パスワードを入力"
                            autoComplete="current-password"
                        />
                        {validationErrors.password && (
                            <p style={{ color: '#ff6b6b', fontSize: '12px', margin: '4px 0 0 0' }}>
                                {validationErrors.password}
                            </p>
                        )}
                    </div>

                    {/* 認証エラー表示 */}
                    {error && (
                        <div style={{
                            padding: '12px',
                            marginBottom: '16px',
                            backgroundColor: '#ffebee',
                            border: '1px solid #ffcdd2',
                            borderRadius: '4px',
                            color: '#c62828',
                            fontSize: '14px',
                        }}>
                            ❌ {error}
                        </div>
                    )}

                    {/* ログインボタン */}
                    <button
                        type="submit"
                        disabled={isLoading}
                        style={{
                            width: '100%',
                            padding: '12px',
                            backgroundColor: isLoading ? '#ccc' : '#007bff',
                            color: 'white',
                            border: 'none',
                            borderRadius: '4px',
                            fontSize: '16px',
                            fontWeight: 'bold',
                            cursor: isLoading ? 'not-allowed' : 'pointer',
                            transition: 'background-color 0.2s',
                        }}
                    >
                        {isLoading ? '🔄 ログイン中...' : 'ログイン'}
                    </button>
                </form>

                {/* テスト用認証情報の表示 */}
                <div style={{
                    marginTop: '20px',
                    padding: '12px',
                    backgroundColor: '#e8f5e8',
                    border: '1px solid #c8e6c9',
                    borderRadius: '4px',
                    fontSize: '12px',
                }}>
                    <p style={{ margin: '0 0 8px 0', fontWeight: 'bold' }}>💡 テスト用認証情報:</p>
                    <p style={{ margin: '0' }}>ユーザー名: test_admin</p>
                    <p style={{ margin: '0' }}>パスワード: password</p>
                </div>
            </div>
        </div>
    );
};