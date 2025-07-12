import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
export type MakeEmpty<T extends { [key: string]: unknown }, K extends keyof T> = { [_ in K]?: never };
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
const defaultOptions = {} as const;
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: { input: string; output: string; }
  String: { input: string; output: string; }
  Boolean: { input: boolean; output: boolean; }
  Int: { input: number; output: number; }
  Float: { input: number; output: number; }
  DateTime: { input: any; output: any; }
};

export type AuthPayload = {
  __typename?: 'AuthPayload';
  expiresAt: Scalars['DateTime']['output'];
  refreshToken: Scalars['String']['output'];
  token: Scalars['String']['output'];
  user: User;
};

export type CreateTestSuiteInput = {
  description: InputMaybe<Scalars['String']['input']>;
  estimatedEndDate: Scalars['DateTime']['input'];
  estimatedStartDate: Scalars['DateTime']['input'];
  name: Scalars['String']['input'];
  requireEffortComment: InputMaybe<Scalars['Boolean']['input']>;
};

export type CreateUserInput = {
  password: Scalars['String']['input'];
  role: Scalars['String']['input'];
  username: Scalars['String']['input'];
};

export type Mutation = {
  __typename?: 'Mutation';
  changePassword: Scalars['Boolean']['output'];
  createTestSuite: TestSuite;
  createUser: User;
  deleteUser: Scalars['Boolean']['output'];
  login: AuthPayload;
  logout: Scalars['Boolean']['output'];
  refreshToken: AuthPayload;
  resetPassword: Scalars['Boolean']['output'];
  updateTestSuite: TestSuite;
  updateTestSuiteStatus: TestSuite;
  updateUser: User;
};


export type MutationChangePasswordArgs = {
  newPassword: Scalars['String']['input'];
  oldPassword: Scalars['String']['input'];
};


export type MutationCreateTestSuiteArgs = {
  input: CreateTestSuiteInput;
};


export type MutationCreateUserArgs = {
  input: CreateUserInput;
};


export type MutationDeleteUserArgs = {
  userId: Scalars['ID']['input'];
};


export type MutationLoginArgs = {
  password: Scalars['String']['input'];
  username: Scalars['String']['input'];
};


export type MutationLogoutArgs = {
  refreshToken: Scalars['String']['input'];
};


export type MutationRefreshTokenArgs = {
  refreshToken: Scalars['String']['input'];
};


export type MutationResetPasswordArgs = {
  newPassword: Scalars['String']['input'];
  userId: Scalars['ID']['input'];
};


export type MutationUpdateTestSuiteArgs = {
  id: Scalars['ID']['input'];
  input: UpdateTestSuiteInput;
};


export type MutationUpdateTestSuiteStatusArgs = {
  id: Scalars['ID']['input'];
  status: SuiteStatus;
};


export type MutationUpdateUserArgs = {
  input: UpdateUserInput;
  userId: Scalars['ID']['input'];
};

export type PageInfo = {
  __typename?: 'PageInfo';
  endCursor: Maybe<Scalars['String']['output']>;
  hasNextPage: Scalars['Boolean']['output'];
  hasPreviousPage: Scalars['Boolean']['output'];
  startCursor: Maybe<Scalars['String']['output']>;
};

export enum Priority {
  Critical = 'CRITICAL',
  High = 'HIGH',
  Low = 'LOW',
  Medium = 'MEDIUM'
}

export type Query = {
  __typename?: 'Query';
  adminData: Scalars['String']['output'];
  managerData: Scalars['String']['output'];
  me: User;
  testSuite: Maybe<TestSuite>;
  testSuites: TestSuiteConnection;
  testerData: Scalars['String']['output'];
  user: Maybe<User>;
  users: Array<User>;
};


export type QueryTestSuiteArgs = {
  id: Scalars['ID']['input'];
};


export type QueryTestSuitesArgs = {
  page: InputMaybe<Scalars['Int']['input']>;
  pageSize: InputMaybe<Scalars['Int']['input']>;
  status: InputMaybe<SuiteStatus>;
};


export type QueryUserArgs = {
  id: Scalars['ID']['input'];
};

export type Subscription = {
  __typename?: 'Subscription';
  testSuiteStatusChanged: TestSuite;
};

export enum SuiteStatus {
  Completed = 'COMPLETED',
  InProgress = 'IN_PROGRESS',
  Preparation = 'PREPARATION',
  Suspended = 'SUSPENDED'
}

export type TestCase = {
  __typename?: 'TestCase';
  actualEffort: Maybe<Scalars['Float']['output']>;
  createdAt: Scalars['DateTime']['output'];
  delayDays: Maybe<Scalars['Int']['output']>;
  description: Maybe<Scalars['String']['output']>;
  groupId: Scalars['ID']['output'];
  id: Scalars['ID']['output'];
  isDelayed: Scalars['Boolean']['output'];
  plannedEffort: Maybe<Scalars['Float']['output']>;
  priority: Priority;
  status: TestStatus;
  title: Scalars['String']['output'];
  updatedAt: Scalars['DateTime']['output'];
};

export type TestGroup = {
  __typename?: 'TestGroup';
  cases: Maybe<Array<TestCase>>;
  createdAt: Scalars['DateTime']['output'];
  description: Maybe<Scalars['String']['output']>;
  displayOrder: Scalars['Int']['output'];
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  status: SuiteStatus;
  suiteId: Scalars['ID']['output'];
  updatedAt: Scalars['DateTime']['output'];
};

export enum TestStatus {
  Completed = 'COMPLETED',
  Created = 'CREATED',
  Fixing = 'FIXING',
  Retesting = 'RETESTING',
  Reviewing = 'REVIEWING',
  ReviewWaiting = 'REVIEW_WAITING',
  Testing = 'TESTING'
}

export type TestSuite = {
  __typename?: 'TestSuite';
  createdAt: Scalars['DateTime']['output'];
  description: Maybe<Scalars['String']['output']>;
  estimatedEndDate: Scalars['DateTime']['output'];
  estimatedStartDate: Scalars['DateTime']['output'];
  groups: Maybe<Array<TestGroup>>;
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  progress: Scalars['Float']['output'];
  requireEffortComment: Scalars['Boolean']['output'];
  status: SuiteStatus;
  updatedAt: Scalars['DateTime']['output'];
};

export type TestSuiteConnection = {
  __typename?: 'TestSuiteConnection';
  edges: Array<TestSuiteEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type TestSuiteEdge = {
  __typename?: 'TestSuiteEdge';
  cursor: Scalars['String']['output'];
  node: TestSuite;
};

export type UpdateTestSuiteInput = {
  description: InputMaybe<Scalars['String']['input']>;
  estimatedEndDate: InputMaybe<Scalars['DateTime']['input']>;
  estimatedStartDate: InputMaybe<Scalars['DateTime']['input']>;
  name: InputMaybe<Scalars['String']['input']>;
  requireEffortComment: InputMaybe<Scalars['Boolean']['input']>;
};

export type UpdateUserInput = {
  role: Scalars['String']['input'];
  username: Scalars['String']['input'];
};

export type User = {
  __typename?: 'User';
  createdAt: Scalars['DateTime']['output'];
  id: Scalars['ID']['output'];
  lastLoginAt: Maybe<Scalars['DateTime']['output']>;
  role: Scalars['String']['output'];
  updatedAt: Scalars['DateTime']['output'];
  username: Scalars['String']['output'];
};

export type LoginMutationVariables = Exact<{
  username: Scalars['String']['input'];
  password: Scalars['String']['input'];
}>;


export type LoginMutation = { __typename?: 'Mutation', login: { __typename?: 'AuthPayload', token: string, refreshToken: string, expiresAt: any, user: { __typename?: 'User', id: string, username: string, role: string, createdAt: any, updatedAt: any, lastLoginAt: any | null } } };

export type LogoutMutationVariables = Exact<{
  refreshToken: Scalars['String']['input'];
}>;


export type LogoutMutation = { __typename?: 'Mutation', logout: boolean };

export type MeQueryVariables = Exact<{ [key: string]: never; }>;


export type MeQuery = { __typename?: 'Query', me: { __typename?: 'User', id: string, username: string, role: string, createdAt: any, updatedAt: any, lastLoginAt: any | null } };

export type CreateTestSuiteMutationVariables = Exact<{
  input: CreateTestSuiteInput;
}>;


export type CreateTestSuiteMutation = { __typename?: 'Mutation', createTestSuite: { __typename?: 'TestSuite', id: string, name: string, description: string | null, status: SuiteStatus, estimatedStartDate: any, estimatedEndDate: any, requireEffortComment: boolean, progress: number, createdAt: any, updatedAt: any } };

export type TestSuiteFieldsFragment = { __typename?: 'TestSuite', id: string, name: string, description: string | null, status: SuiteStatus, estimatedStartDate: any, estimatedEndDate: any, requireEffortComment: boolean, progress: number, createdAt: any, updatedAt: any };

export type GetTestSuiteListQueryVariables = Exact<{
  status: InputMaybe<SuiteStatus>;
  page: InputMaybe<Scalars['Int']['input']>;
  pageSize: InputMaybe<Scalars['Int']['input']>;
}>;


export type GetTestSuiteListQuery = { __typename?: 'Query', testSuites: { __typename?: 'TestSuiteConnection', totalCount: number, edges: Array<{ __typename?: 'TestSuiteEdge', cursor: string, node: { __typename?: 'TestSuite', id: string, name: string, description: string | null, status: SuiteStatus, estimatedStartDate: any, estimatedEndDate: any, requireEffortComment: boolean, progress: number, createdAt: any, updatedAt: any } }>, pageInfo: { __typename?: 'PageInfo', hasNextPage: boolean, hasPreviousPage: boolean, startCursor: string | null, endCursor: string | null } } };

export type GetTestSuiteDetailQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type GetTestSuiteDetailQuery = { __typename?: 'Query', testSuite: { __typename?: 'TestSuite', id: string, name: string, description: string | null, status: SuiteStatus, estimatedStartDate: any, estimatedEndDate: any, requireEffortComment: boolean, progress: number, createdAt: any, updatedAt: any, groups: Array<{ __typename?: 'TestGroup', id: string, name: string, description: string | null, displayOrder: number, status: SuiteStatus, createdAt: any, updatedAt: any }> | null } | null };

export const TestSuiteFieldsFragmentDoc = gql`
    fragment TestSuiteFields on TestSuite {
  id
  name
  description
  status
  estimatedStartDate
  estimatedEndDate
  requireEffortComment
  progress
  createdAt
  updatedAt
}
    `;
export const LoginDocument = gql`
    mutation Login($username: String!, $password: String!) {
  login(username: $username, password: $password) {
    token
    refreshToken
    user {
      id
      username
      role
      createdAt
      updatedAt
      lastLoginAt
    }
    expiresAt
  }
}
    `;
export type LoginMutationFn = Apollo.MutationFunction<LoginMutation, LoginMutationVariables>;

/**
 * __useLoginMutation__
 *
 * To run a mutation, you first call `useLoginMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useLoginMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [loginMutation, { data, loading, error }] = useLoginMutation({
 *   variables: {
 *      username: // value for 'username'
 *      password: // value for 'password'
 *   },
 * });
 */
export function useLoginMutation(baseOptions?: Apollo.MutationHookOptions<LoginMutation, LoginMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<LoginMutation, LoginMutationVariables>(LoginDocument, options);
      }
export type LoginMutationHookResult = ReturnType<typeof useLoginMutation>;
export type LoginMutationResult = Apollo.MutationResult<LoginMutation>;
export type LoginMutationOptions = Apollo.BaseMutationOptions<LoginMutation, LoginMutationVariables>;
export const LogoutDocument = gql`
    mutation Logout($refreshToken: String!) {
  logout(refreshToken: $refreshToken)
}
    `;
export type LogoutMutationFn = Apollo.MutationFunction<LogoutMutation, LogoutMutationVariables>;

/**
 * __useLogoutMutation__
 *
 * To run a mutation, you first call `useLogoutMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useLogoutMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [logoutMutation, { data, loading, error }] = useLogoutMutation({
 *   variables: {
 *      refreshToken: // value for 'refreshToken'
 *   },
 * });
 */
export function useLogoutMutation(baseOptions?: Apollo.MutationHookOptions<LogoutMutation, LogoutMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<LogoutMutation, LogoutMutationVariables>(LogoutDocument, options);
      }
export type LogoutMutationHookResult = ReturnType<typeof useLogoutMutation>;
export type LogoutMutationResult = Apollo.MutationResult<LogoutMutation>;
export type LogoutMutationOptions = Apollo.BaseMutationOptions<LogoutMutation, LogoutMutationVariables>;
export const MeDocument = gql`
    query Me {
  me {
    id
    username
    role
    createdAt
    updatedAt
    lastLoginAt
  }
}
    `;

/**
 * __useMeQuery__
 *
 * To run a query within a React component, call `useMeQuery` and pass it any options that fit your needs.
 * When your component renders, `useMeQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useMeQuery({
 *   variables: {
 *   },
 * });
 */
export function useMeQuery(baseOptions?: Apollo.QueryHookOptions<MeQuery, MeQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<MeQuery, MeQueryVariables>(MeDocument, options);
      }
export function useMeLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<MeQuery, MeQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<MeQuery, MeQueryVariables>(MeDocument, options);
        }
export function useMeSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<MeQuery, MeQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<MeQuery, MeQueryVariables>(MeDocument, options);
        }
export type MeQueryHookResult = ReturnType<typeof useMeQuery>;
export type MeLazyQueryHookResult = ReturnType<typeof useMeLazyQuery>;
export type MeSuspenseQueryHookResult = ReturnType<typeof useMeSuspenseQuery>;
export type MeQueryResult = Apollo.QueryResult<MeQuery, MeQueryVariables>;
export const CreateTestSuiteDocument = gql`
    mutation CreateTestSuite($input: CreateTestSuiteInput!) {
  createTestSuite(input: $input) {
    id
    name
    description
    status
    estimatedStartDate
    estimatedEndDate
    requireEffortComment
    progress
    createdAt
    updatedAt
  }
}
    `;
export type CreateTestSuiteMutationFn = Apollo.MutationFunction<CreateTestSuiteMutation, CreateTestSuiteMutationVariables>;

/**
 * __useCreateTestSuiteMutation__
 *
 * To run a mutation, you first call `useCreateTestSuiteMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateTestSuiteMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createTestSuiteMutation, { data, loading, error }] = useCreateTestSuiteMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useCreateTestSuiteMutation(baseOptions?: Apollo.MutationHookOptions<CreateTestSuiteMutation, CreateTestSuiteMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateTestSuiteMutation, CreateTestSuiteMutationVariables>(CreateTestSuiteDocument, options);
      }
export type CreateTestSuiteMutationHookResult = ReturnType<typeof useCreateTestSuiteMutation>;
export type CreateTestSuiteMutationResult = Apollo.MutationResult<CreateTestSuiteMutation>;
export type CreateTestSuiteMutationOptions = Apollo.BaseMutationOptions<CreateTestSuiteMutation, CreateTestSuiteMutationVariables>;
export const GetTestSuiteListDocument = gql`
    query GetTestSuiteList($status: SuiteStatus, $page: Int, $pageSize: Int) {
  testSuites(status: $status, page: $page, pageSize: $pageSize) {
    edges {
      node {
        ...TestSuiteFields
      }
      cursor
    }
    pageInfo {
      hasNextPage
      hasPreviousPage
      startCursor
      endCursor
    }
    totalCount
  }
}
    ${TestSuiteFieldsFragmentDoc}`;

/**
 * __useGetTestSuiteListQuery__
 *
 * To run a query within a React component, call `useGetTestSuiteListQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetTestSuiteListQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetTestSuiteListQuery({
 *   variables: {
 *      status: // value for 'status'
 *      page: // value for 'page'
 *      pageSize: // value for 'pageSize'
 *   },
 * });
 */
export function useGetTestSuiteListQuery(baseOptions?: Apollo.QueryHookOptions<GetTestSuiteListQuery, GetTestSuiteListQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetTestSuiteListQuery, GetTestSuiteListQueryVariables>(GetTestSuiteListDocument, options);
      }
export function useGetTestSuiteListLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetTestSuiteListQuery, GetTestSuiteListQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetTestSuiteListQuery, GetTestSuiteListQueryVariables>(GetTestSuiteListDocument, options);
        }
export function useGetTestSuiteListSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<GetTestSuiteListQuery, GetTestSuiteListQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<GetTestSuiteListQuery, GetTestSuiteListQueryVariables>(GetTestSuiteListDocument, options);
        }
export type GetTestSuiteListQueryHookResult = ReturnType<typeof useGetTestSuiteListQuery>;
export type GetTestSuiteListLazyQueryHookResult = ReturnType<typeof useGetTestSuiteListLazyQuery>;
export type GetTestSuiteListSuspenseQueryHookResult = ReturnType<typeof useGetTestSuiteListSuspenseQuery>;
export type GetTestSuiteListQueryResult = Apollo.QueryResult<GetTestSuiteListQuery, GetTestSuiteListQueryVariables>;
export const GetTestSuiteDetailDocument = gql`
    query GetTestSuiteDetail($id: ID!) {
  testSuite(id: $id) {
    ...TestSuiteFields
    groups {
      id
      name
      description
      displayOrder
      status
      createdAt
      updatedAt
    }
  }
}
    ${TestSuiteFieldsFragmentDoc}`;

/**
 * __useGetTestSuiteDetailQuery__
 *
 * To run a query within a React component, call `useGetTestSuiteDetailQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetTestSuiteDetailQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetTestSuiteDetailQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useGetTestSuiteDetailQuery(baseOptions: Apollo.QueryHookOptions<GetTestSuiteDetailQuery, GetTestSuiteDetailQueryVariables> & ({ variables: GetTestSuiteDetailQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetTestSuiteDetailQuery, GetTestSuiteDetailQueryVariables>(GetTestSuiteDetailDocument, options);
      }
export function useGetTestSuiteDetailLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetTestSuiteDetailQuery, GetTestSuiteDetailQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetTestSuiteDetailQuery, GetTestSuiteDetailQueryVariables>(GetTestSuiteDetailDocument, options);
        }
export function useGetTestSuiteDetailSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<GetTestSuiteDetailQuery, GetTestSuiteDetailQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<GetTestSuiteDetailQuery, GetTestSuiteDetailQueryVariables>(GetTestSuiteDetailDocument, options);
        }
export type GetTestSuiteDetailQueryHookResult = ReturnType<typeof useGetTestSuiteDetailQuery>;
export type GetTestSuiteDetailLazyQueryHookResult = ReturnType<typeof useGetTestSuiteDetailLazyQuery>;
export type GetTestSuiteDetailSuspenseQueryHookResult = ReturnType<typeof useGetTestSuiteDetailSuspenseQuery>;
export type GetTestSuiteDetailQueryResult = Apollo.QueryResult<GetTestSuiteDetailQuery, GetTestSuiteDetailQueryVariables>;