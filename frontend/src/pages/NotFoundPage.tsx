// src/pages/NotFoundPage.tsx

import React from 'react';
import { Link } from 'react-router-dom';

export const NotFoundPage: React.FC = () => (
    <div style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        backgroundColor: '#f5f5f5',
        flexDirection: 'column',
        gap: '24px',
        textAlign: 'center',
        padding: '20px'
    }}>
        <div style={{ fontSize: '72px' }}>🔍</div>

        <div>
            <h1 style={{ margin: '0 0 16px 0', fontSize: '48px', color: '#666' }}>404</h1>
            <h2 style={{ margin: '0 0 8px 0', color: '#333' }}>ページが見つかりません</h2>
            <p style={{ margin: '0', color: '#666' }}>
                お探しのページは存在しないか、移動された可能性があります。
            </p>
        </div>

        <div style={{ display: 'flex', gap: '16px' }}>
            <Link
                to="/"
                style={{
                    padding: '12px 24px',
                    backgroundColor: '#007bff',
                    color: 'white',
                    textDecoration: 'none',
                    borderRadius: '4px',
                    fontSize: '16px'
                }}
            >
                🏠 ダッシュボードに戻る
            </Link>

            <Link
                to="/test-suites"
                style={{
                    padding: '12px 24px',
                    backgroundColor: '#6c757d',
                    color: 'white',
                    textDecoration: 'none',
                    borderRadius: '4px',
                    fontSize: '16px'
                }}
            >
                📋 テストスイート一覧
            </Link>
        </div>
    </div>
);