// src/components/MainNavigation.tsx

import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

export const MainNavigation: React.FC = () => {
    const { user, logout } = useAuth();
    const location = useLocation();

    const handleLogout = async () => {
        try {
            await logout();
        } catch (error) {
            console.error('ログアウトエラー:', error);
        }
    };

    return (
        <header style={{
            backgroundColor: '#343a40',
            color: 'white',
            padding: '12px 20px',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: '0'
        }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '24px' }}>
                <h1 style={{ margin: '0', fontSize: '20px' }}>テスト管理システム</h1>

                <nav style={{ display: 'flex', gap: '16px' }}>
                    <Link
                        to="/"
                        style={{
                            padding: '8px 16px',
                            backgroundColor: location.pathname === '/' ? '#007bff' : 'transparent',
                            color: 'white',
                            border: '1px solid #6c757d',
                            borderRadius: '4px',
                            textDecoration: 'none',
                            fontSize: '14px',
                            display: 'inline-block'
                        }}
                    >
                        🏠 ダッシュボード
                    </Link>

                    <Link
                        to="/test-suites"
                        style={{
                            padding: '8px 16px',
                            backgroundColor: location.pathname === '/test-suites' ? '#007bff' : 'transparent',
                            color: 'white',
                            border: '1px solid #6c757d',
                            borderRadius: '4px',
                            textDecoration: 'none',
                            fontSize: '14px',
                            display: 'inline-block'
                        }}
                    >
                        📋 テストスイート
                    </Link>
                </nav>
            </div>

            <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
                <span style={{ fontSize: '14px' }}>
                    {user?.username} ({user?.role})
                </span>
                <button
                    onClick={handleLogout}
                    style={{
                        padding: '6px 12px',
                        backgroundColor: '#dc3545',
                        color: 'white',
                        border: 'none',
                        borderRadius: '4px',
                        cursor: 'pointer',
                        fontSize: '12px'
                    }}
                >
                    ログアウト
                </button>
            </div>
        </header>
    );
};