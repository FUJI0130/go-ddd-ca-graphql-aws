// src/contexts/AuthContext.tsx

import React, { createContext, useContext, useReducer, useEffect } from 'react';
import { useMutation, useLazyQuery } from '@apollo/client';
import {
    AuthState,
    AuthContextType,
    LoginCredentials,
    AuthUser
} from '../types/auth';
import {
    LoginDocument,
    LogoutDocument,
    MeDocument
} from '../generated/graphql';

// 認証状態の初期値
const initialState: AuthState = {
    isAuthenticated: false,
    isLoading: true,
    user: null,
    error: null,
};

// 認証アクション型
type AuthAction =
    | { type: 'AUTH_START' }
    | { type: 'AUTH_SUCCESS'; payload: AuthUser }
    | { type: 'AUTH_FAILURE'; payload: string }
    | { type: 'AUTH_LOGOUT' }
    | { type: 'AUTH_RESET_ERROR' }
    | { type: 'AUTH_SET_LOADING'; payload: boolean };

// 認証状態のReducer
const authReducer = (state: AuthState, action: AuthAction): AuthState => {
    switch (action.type) {
        case 'AUTH_START':
            return {
                ...state,
                isLoading: true,
                error: null,
            };

        case 'AUTH_SUCCESS':
            return {
                ...state,
                isAuthenticated: true,
                isLoading: false,
                user: action.payload,
                error: null,
            };

        case 'AUTH_FAILURE':
            return {
                ...state,
                isAuthenticated: false,
                isLoading: false,
                user: null,
                error: action.payload,
            };

        case 'AUTH_LOGOUT':
            return {
                ...state,
                isAuthenticated: false,
                isLoading: false,
                user: null,
                error: null,
            };

        case 'AUTH_RESET_ERROR':
            return {
                ...state,
                error: null,
            };

        case 'AUTH_SET_LOADING':
            return {
                ...state,
                isLoading: action.payload,
            };

        default:
            return state;
    }
};

// 認証コンテキストの作成
const AuthContext = createContext<AuthContextType | undefined>(undefined);

// 認証プロバイダーコンポーネント
export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const [state, dispatch] = useReducer(authReducer, initialState);

    // GraphQLミューテーション・クエリのフック
    const [loginMutation] = useMutation(LoginDocument);
    const [logoutMutation] = useMutation(LogoutDocument);
    const [checkMe, { data: meData }] = useLazyQuery(MeDocument, {
        errorPolicy: 'ignore', // 認証エラーは無視（未ログイン状態として扱う）
    });

    // ログイン機能
    const login = async (credentials: LoginCredentials): Promise<void> => {
        try {
            dispatch({ type: 'AUTH_START' });

            const { data } = await loginMutation({
                variables: {
                    username: credentials.username,
                    password: credentials.password,
                },
            });

            if (data?.login?.user) {
                // Cookie認証のため、トークンはサーバー側で設定済み
                dispatch({
                    type: 'AUTH_SUCCESS',
                    payload: data.login.user as AuthUser
                });
            } else {
                throw new Error('ログインに失敗しました');
            }
        } catch (error) {
            const errorMessage = error instanceof Error ? error.message : 'ログインエラーが発生しました';
            dispatch({ type: 'AUTH_FAILURE', payload: errorMessage });
            throw error;
        }
    };

    // ログアウト機能
    // ログアウト機能（修正版）
    const logout = async (): Promise<void> => {
        try {
            // リフレッシュトークンはCookieから自動送信される
            const { data } = await logoutMutation({
                variables: {
                    refreshToken: '', // Cookieから自動取得されるため空文字
                },
            });

            // Boolean型の戻り値を確認（オプション）
            if (data?.logout === true) {
                console.log('ログアウト成功');
            }

            dispatch({ type: 'AUTH_LOGOUT' });
        } catch (error) {
            console.error('ログアウトエラー:', error);
            // ログアウトはエラーが発生してもローカル状態をクリア
            dispatch({ type: 'AUTH_LOGOUT' });
        }
    };

    // エラー状態リセット
    const resetAuthError = (): void => {
        dispatch({ type: 'AUTH_RESET_ERROR' });
    };

    // 認証状態確認（Cookie認証の確認）
    const checkAuthStatus = async (): Promise<void> => {
        try {
            dispatch({ type: 'AUTH_SET_LOADING', payload: true });

            const result = await checkMe({
                fetchPolicy: 'network-only',  // キャッシュ回避
                errorPolicy: 'ignore'         // 認証エラー許容
            });

            if (result.data?.me) {
                dispatch({
                    type: 'AUTH_SUCCESS',
                    payload: result.data.me as AuthUser
                });
            } else {
                dispatch({ type: 'AUTH_LOGOUT' });
            }
        } catch (error) {
            console.error('認証確認エラー:', error);
            dispatch({ type: 'AUTH_LOGOUT' });
        } finally {
            // ✅ 確実なローディング解除
            dispatch({ type: 'AUTH_SET_LOADING', payload: false });
        }
    };

    // 初回ロード時の認証状態確認
    useEffect(() => {
        checkAuthStatus();
    }, []);

    // meDataの変更を監視
    useEffect(() => {
        if (meData?.me) {
            dispatch({
                type: 'AUTH_SUCCESS',
                payload: meData.me as AuthUser
            });
        }
    }, [meData]);

    const contextValue: AuthContextType = {
        ...state,
        login,
        logout,
        resetAuthError,
        checkAuthStatus,
    };

    return (
        <AuthContext.Provider value={contextValue}>
            {children}
        </AuthContext.Provider>
    );
};

// 認証コンテキストを使用するカスタムフック
export const useAuth = (): AuthContextType => {
    const context = useContext(AuthContext);
    if (context === undefined) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
};