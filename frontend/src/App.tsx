// src/App.tsx - React Router統合版

import React from 'react';
import { AuthProvider } from './contexts/AuthContext';
import { AppRouter } from './routes/AppRouter';
import './App.css';

// メインアプリコンポーネント
const App: React.FC = () => {
    return (
        <AuthProvider>
            <AppRouter />
        </AuthProvider>
    );
};

export default App;