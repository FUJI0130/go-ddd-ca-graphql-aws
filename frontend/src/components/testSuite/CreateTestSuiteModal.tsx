// src/components/testSuite/CreateTestSuiteModal.tsx

import React, { useState, ChangeEvent, FormEvent } from 'react';
import { useCreateTestSuite } from '../../hooks/useTestSuites';
// Generated型を使用
import { CreateTestSuiteInput } from '../../generated/graphql';

interface CreateTestSuiteModalProps {
  onClose: () => void;
  onSuccess: () => void;
}

// 日付フォーマット変換ヘルパー関数
const formatDateForGraphQL = (dateString: string): string => {
  if (!dateString) return '';

  try {
    // YYYY-MM-DD を ISO文字列（YYYY-MM-DDTHH:mm:ss.sssZ）に変換
    const date = new Date(dateString + 'T00:00:00.000Z');
    return date.toISOString();
  } catch (error) {
    console.error('日付変換エラー:', error);
    return dateString; // フォールバック
  }
};

export const CreateTestSuiteModal: React.FC<CreateTestSuiteModalProps> = ({
  onClose,
  onSuccess
}) => {
  const { createTestSuite, loading, error } = useCreateTestSuite();

  // フォーム用の内部状態（HTML date input用にstring型を維持）
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    estimatedStartDate: '',     // YYYY-MM-DD 形式で保持
    estimatedEndDate: '',       // YYYY-MM-DD 形式で保持
    requireEffortComment: false
  });

  const [validationErrors, setValidationErrors] = useState<{
    name?: string;
    estimatedStartDate?: string;
    estimatedEndDate?: string;
  }>({});

  const validateForm = (): boolean => {
    const errors: typeof validationErrors = {};

    if (!formData.name.trim()) {
      errors.name = '名前を入力してください';
    }

    if (!formData.estimatedStartDate) {
      errors.estimatedStartDate = '開始予定日を選択してください';
    }

    if (!formData.estimatedEndDate) {
      errors.estimatedEndDate = '終了予定日を選択してください';
    }

    if (formData.estimatedStartDate && formData.estimatedEndDate) {
      if (new Date(formData.estimatedStartDate) >= new Date(formData.estimatedEndDate)) {
        errors.estimatedEndDate = '終了予定日は開始予定日より後の日付を選択してください';
      }
    }

    setValidationErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    try {
      // GraphQL用のデータに変換（日付をISO文字列に変換）
      const graphqlInput: CreateTestSuiteInput = {
        name: formData.name.trim(),
        description: formData.description?.trim() || null,
        estimatedStartDate: formatDateForGraphQL(formData.estimatedStartDate), // ✅ ISO文字列に変換
        estimatedEndDate: formatDateForGraphQL(formData.estimatedEndDate),     // ✅ ISO文字列に変換
        requireEffortComment: formData.requireEffortComment || false
      };

      console.log('送信データ:', graphqlInput); // デバッグ用ログ

      await createTestSuite(graphqlInput);
      onSuccess();
    } catch (error) {
      console.error('テストスイート作成エラー:', error);
    }
  };

  const handleInputChange = (
    e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) => {
    const { name, value, type } = e.target;
    const target = e.target as HTMLInputElement;

    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? target.checked : value
    }));

    // バリデーションエラーをクリア
    if (validationErrors[name as keyof typeof validationErrors]) {
      setValidationErrors(prev => ({
        ...prev,
        [name]: undefined
      }));
    }
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
        width: '500px',
        maxWidth: '90vw',
        maxHeight: '90vh',
        overflow: 'auto'
      }}>
        <h3 style={{ margin: '0 0 20px 0' }}>新しいテストスイート作成</h3>

        <form onSubmit={handleSubmit}>
          {/* 名前 */}
          <div style={{ marginBottom: '16px' }}>
            <label style={{ display: 'block', marginBottom: '8px', fontWeight: 'bold' }}>
              名前 <span style={{ color: '#dc3545' }}>*</span>
            </label>
            <input
              type="text"
              name="name"
              value={formData.name}
              onChange={handleInputChange}
              placeholder="テストスイート名を入力"
              style={{
                width: '100%',
                padding: '8px',
                border: `1px solid ${validationErrors.name ? '#dc3545' : '#ddd'}`,
                borderRadius: '4px'
              }}
            />
            {validationErrors.name && (
              <p style={{ color: '#dc3545', fontSize: '12px', margin: '4px 0 0 0' }}>
                {validationErrors.name}
              </p>
            )}
          </div>

          {/* 説明 */}
          <div style={{ marginBottom: '16px' }}>
            <label style={{ display: 'block', marginBottom: '8px', fontWeight: 'bold' }}>
              説明
            </label>
            <textarea
              name="description"
              value={formData.description || ''}
              onChange={handleInputChange}
              placeholder="テストスイートの説明を入力"
              rows={3}
              style={{
                width: '100%',
                padding: '8px',
                border: '1px solid #ddd',
                borderRadius: '4px',
                resize: 'vertical'
              }}
            />
          </div>

          {/* 日付設定 */}
          <div style={{
            display: 'grid',
            gridTemplateColumns: '1fr 1fr',
            gap: '16px',
            marginBottom: '16px'
          }}>
            <div>
              <label style={{ display: 'block', marginBottom: '8px', fontWeight: 'bold' }}>
                開始予定日 <span style={{ color: '#dc3545' }}>*</span>
              </label>
              <input
                type="date"
                name="estimatedStartDate"
                value={formData.estimatedStartDate}
                onChange={handleInputChange}
                style={{
                  width: '100%',
                  padding: '8px',
                  border: `1px solid ${validationErrors.estimatedStartDate ? '#dc3545' : '#ddd'}`,
                  borderRadius: '4px'
                }}
              />
              {validationErrors.estimatedStartDate && (
                <p style={{ color: '#dc3545', fontSize: '12px', margin: '4px 0 0 0' }}>
                  {validationErrors.estimatedStartDate}
                </p>
              )}
            </div>

            <div>
              <label style={{ display: 'block', marginBottom: '8px', fontWeight: 'bold' }}>
                終了予定日 <span style={{ color: '#dc3545' }}>*</span>
              </label>
              <input
                type="date"
                name="estimatedEndDate"
                value={formData.estimatedEndDate}
                onChange={handleInputChange}
                style={{
                  width: '100%',
                  padding: '8px',
                  border: `1px solid ${validationErrors.estimatedEndDate ? '#dc3545' : '#ddd'}`,
                  borderRadius: '4px'
                }}
              />
              {validationErrors.estimatedEndDate && (
                <p style={{ color: '#dc3545', fontSize: '12px', margin: '4px 0 0 0' }}>
                  {validationErrors.estimatedEndDate}
                </p>
              )}
            </div>
          </div>

          {/* 工数コメント要求 */}
          <div style={{ marginBottom: '20px' }}>
            <label style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
              <input
                type="checkbox"
                name="requireEffortComment"
                checked={formData.requireEffortComment || false}
                onChange={handleInputChange}
              />
              <span style={{ fontWeight: 'bold' }}>工数入力時にコメントを必須とする</span>
            </label>
          </div>

          {/* エラー表示（改善版） */}
          {error && (
            <div style={{
              padding: '12px',
              marginBottom: '16px',
              backgroundColor: '#f8d7da',
              border: '1px solid #f5c6cb',
              borderRadius: '4px',
              color: '#721c24',
              fontSize: '14px'
            }}>
              <strong>作成に失敗しました</strong><br />
              {(error as any)?.message || 'エラーが発生しました。入力内容を確認してください。'}
            </div>
          )}

          {/* ボタン群 */}
          <div style={{
            display: 'flex',
            justifyContent: 'flex-end',
            gap: '8px'
          }}>
            <button
              type="button"
              onClick={onClose}
              disabled={loading}
              style={{
                padding: '8px 16px',
                backgroundColor: 'transparent',
                color: '#666',
                border: '1px solid #ddd',
                borderRadius: '4px',
                cursor: loading ? 'not-allowed' : 'pointer'
              }}
            >
              キャンセル
            </button>

            <button
              type="submit"
              disabled={loading}
              style={{
                padding: '8px 16px',
                backgroundColor: loading ? '#ccc' : '#28a745',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                cursor: loading ? 'not-allowed' : 'pointer'
              }}
            >
              {loading ? '🔄 作成中...' : '✅ 作成'}
            </button>
          </div>
        </form>

        {/* デバッグ情報（開発時のみ表示） */}
        {process.env.NODE_ENV === 'development' && (
          <details style={{ marginTop: '16px', fontSize: '12px', color: '#666' }}>
            <summary>🔧 デバッグ情報</summary>
            <pre style={{ fontSize: '10px', marginTop: '8px' }}>
              {JSON.stringify({
                formData,
                converted: {
                  estimatedStartDate: formatDateForGraphQL(formData.estimatedStartDate),
                  estimatedEndDate: formatDateForGraphQL(formData.estimatedEndDate)
                }
              }, null, 2)}
            </pre>
          </details>
        )}
      </div>
    </div>
  );
};