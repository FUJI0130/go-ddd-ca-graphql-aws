// src/types/testSuite.ts

// Generated型から必要な型をimport
import { SuiteStatus } from '../generated/graphql';

// Generated型をre-exportして統一（型と値を分離してisolatedModules対応）
export type { 
  TestSuite as GeneratedTestSuite,
  TestGroup as GeneratedTestGroup,
  CreateTestSuiteInput,
  UpdateTestSuiteInput
} from '../generated/graphql';

// 値のexport（enum等）
export { SuiteStatus } from '../generated/graphql';

// ローカル固有の型定義（Generated型にないもののみ）
export interface TestCase {
  id: string;
  title: string;
  description: string;
  status: TestCaseStatus;
  priority: TestCasePriority;
  plannedEffort: number;
  actualEffort: number;
  isDelayed: boolean;
  delayDays: number;
  groupId: string;
  createdAt: string;
  updatedAt: string;
}

// ローカル型（Generated型と互換性を保つため調整）
export interface TestSuite {
  id: string;
  name: string;
  description: string; // Generated型のnull許可に対応
  status: SuiteStatus; // Generated型を使用
  estimatedStartDate: string;
  estimatedEndDate: string;
  requireEffortComment: boolean;
  progress: number;
  createdAt: string;
  updatedAt: string;
  groups?: TestGroup[];
}

export interface TestGroup {
  id: string;
  name: string;
  description: string;
  displayOrder: number;
  suiteId: string;
  status: SuiteStatus; // Generated SuiteStatusを使用（TestGroupもSuiteStatusを使用）
  createdAt: string;
  updatedAt: string;
  cases: TestCase[];
}

// ローカル固有のenum（Generated型にないもの）
export enum TestGroupStatus {
  PENDING = 'PENDING',
  IN_PROGRESS = 'IN_PROGRESS',
  COMPLETED = 'COMPLETED'
}

export enum TestCaseStatus {
  NOT_STARTED = 'NOT_STARTED',
  IN_PROGRESS = 'IN_PROGRESS',
  PASSED = 'PASSED',
  FAILED = 'FAILED',
  SKIPPED = 'SKIPPED'
}

export enum TestCasePriority {
  LOW = 'LOW',
  MEDIUM = 'MEDIUM',
  HIGH = 'HIGH',
  CRITICAL = 'CRITICAL'
}

// ページネーション関連型（Generated型を使用）
export interface TestSuiteConnection {
  edges: TestSuiteEdge[];
  pageInfo: PageInfo;
  totalCount: number;
}

export interface TestSuiteEdge {
  node: TestSuite;
  cursor: string;
}

export interface PageInfo {
  hasNextPage: boolean;
  hasPreviousPage: boolean;
  startCursor?: string;
  endCursor?: string;
}

// フィルター・検索関連型（ローカル固有）
export interface TestSuiteFilters {
  status?: SuiteStatus; // Generated SuiteStatusを使用
  search?: string;
  dateFrom?: string;
  dateTo?: string;
}

export interface TestSuiteListState {
  testSuites: TestSuite[];
  loading: boolean;
  error: string | null;
  filters: TestSuiteFilters;
  totalCount: number;
  hasNextPage: boolean;
}