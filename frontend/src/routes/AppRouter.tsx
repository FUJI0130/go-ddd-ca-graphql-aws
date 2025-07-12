// src/routes/AppRouter.tsx - TypeScriptエラー修正版

import React from 'react';
import { BrowserRouter, Routes, Route, } from 'react-router-dom';
import { ProtectedRoute } from './ProtectedRoute';
import { LoginPage } from '../pages/LoginPage';
import { DashboardPage } from '../pages/DashboardPage';
import { TestSuiteListPage } from '../pages/TestSuiteListPage';
import { NotFoundPage } from '../pages/NotFoundPage';

export const AppRouter: React.FC = () => (
    <BrowserRouter>
        <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/" element={
                <ProtectedRoute><DashboardPage /></ProtectedRoute>
            } />
            <Route path="/test-suites" element={
                <ProtectedRoute><TestSuiteListPage /></ProtectedRoute>
            } />
            {/* 404ページとして NotFoundPage を使用 */}
            <Route path="*" element={<NotFoundPage />} />
        </Routes>
    </BrowserRouter>
);