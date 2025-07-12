// src/layouts/MainLayout.tsx

import React from 'react';
import { MainNavigation } from '../components/MainNavigation';

interface MainLayoutProps {
    children: React.ReactNode;
}

export const MainLayout: React.FC<MainLayoutProps> = ({ children }) => (
    <div style={{ minHeight: '100vh', backgroundColor: '#f8f9fa' }}>
        <MainNavigation />
        <main style={{ minHeight: 'calc(100vh - 60px)' }}>
            {children}
        </main>
    </div>
);