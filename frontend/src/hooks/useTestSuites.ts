// src/hooks/useTestSuites.ts

import { useState, useCallback } from 'react';
import { 
  // Generated型・フックをインポート
  useGetTestSuiteListQuery,
  useCreateTestSuiteMutation,
  CreateTestSuiteInput,
  SuiteStatus
} from '../generated/graphql';

// ローカル型（Generated型にないもののみ）
import { 
  TestSuiteFilters
} from '../types/testSuite';

// テストスイート一覧管理フック（拡張版）
export const useTestSuites = (options: { status?: SuiteStatus, page?: number, pageSize?: number } = {}) => {
  const [filters, setFilters] = useState<TestSuiteFilters>({
    status: options.status,
    search: '',
    dateFrom: '',
    dateTo: ''
  });

  const { data, loading, error, refetch, fetchMore } = useGetTestSuiteListQuery({
    variables: {
      status: options.status || null,
      page: options.page || 1,
      pageSize: options.pageSize || 10
    },
    fetchPolicy: 'cache-and-network',
    errorPolicy: 'all'
  });

  // Generated TestSuiteをローカル形式に変換
  const testSuites = data?.testSuites?.edges?.map(edge => {
    const generatedSuite = edge.node;
    // Generated型をローカル型に変換（必要なフィールドのみ）
    return {
      ...generatedSuite,
      description: generatedSuite.description || '' // null を空文字に変換
      // groupsフィールドはGetTestSuiteListQueryに含まれないため削除
    };
  }) || [];

  const updateFilters = useCallback((newFilters: Partial<TestSuiteFilters>) => {
    setFilters(prev => ({ ...prev, ...newFilters }));
    // フィルター変更時にrefetch実行
    refetch({
      status: newFilters.status || null,
      page: 1, // フィルター変更時は1ページ目に戻す
      pageSize: options.pageSize || 10
    });
  }, [refetch, options.pageSize]);

  const loadMore = useCallback(() => {
    if (data?.testSuites?.pageInfo?.hasNextPage) {
      fetchMore({
        variables: {
          page: (options.page || 1) + 1,
          pageSize: options.pageSize || 10,
          status: options.status || null
        }
      });
    }
  }, [data, fetchMore, options.page, options.pageSize, options.status]);

  return {
    testSuites,
    totalCount: data?.testSuites?.totalCount || 0,
    hasNextPage: data?.testSuites?.pageInfo?.hasNextPage || false,
    loading,
    error,
    filters,
    updateFilters,
    loadMore,
    refresh: refetch
  };
};

// 個別テストスイート管理フック
export const useTestSuite = (_id: string) => {
  return {
    testSuite: null,
    loading: false,
    error: null,
    refetch: () => Promise.resolve()
  };
};

// テストスイート作成フック（Generated Mutationを使用）
export const useCreateTestSuite = () => {
  const [createTestSuiteMutation, { loading, error }] = useCreateTestSuiteMutation();
  
  const createTestSuite = async (input: CreateTestSuiteInput) => {
    try {
      const result = await createTestSuiteMutation({
        variables: { input }
      });
      return result.data?.createTestSuite;
    } catch (err) {
      throw err;
    }
  };
  
  return {
    createTestSuite,
    loading,
    error
  };
};

// テストスイートステータス更新フック
export const useUpdateTestSuiteStatus = () => {
  const updateStatus = async (id: string, status: SuiteStatus) => {
    console.log('ステータス更新（未実装）:', id, status);
    return null;
  };
  
  return {
    updateStatus,
    loading: false,
    error: null
  };
};