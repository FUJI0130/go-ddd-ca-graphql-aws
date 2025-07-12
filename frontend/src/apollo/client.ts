import { ApolloClient, InMemoryCache, createHttpLink, from } from '@apollo/client';
import { onError } from '@apollo/client/link/error';
import { setContext } from '@apollo/client/link/context';

// GraphQLエンドポイントへのHTTPリンク
// const httpLink = createHttpLink({
//   uri: import.meta.env.VITE_GRAPHQL_API_URL || 'http://localhost:8080/query',
// });
const httpLink = createHttpLink({
  uri: import.meta.env.VITE_GRAPHQL_API_URL || 'http://localhost:8080/query',
  credentials: 'include',
});

// エラーハンドリング
const errorLink = onError(({ graphQLErrors, networkError }) => {
  if (graphQLErrors) {
    graphQLErrors.forEach(({ message, locations, path }) => {
      console.error(
        `[GraphQL error]: Message: ${message}, Location: ${locations}, Path: ${path}`
      );
    });
  }
  if (networkError) {
    console.error(`[Network error]: ${networkError}`);
  }
});

// 認証ヘッダーの追加
const authLink = setContext((_, { headers }) => {
  // Cookieは自動送信されるため認証ヘッダー不要
  return {
    headers: {
      ...headers,
    },
  };
});

// Apollo Clientのインスタンス作成
export const client = new ApolloClient({
  link: from([errorLink, authLink, httpLink]),
  cache: new InMemoryCache(),
  defaultOptions: {
    watchQuery: {
      fetchPolicy: 'cache-and-network',
    },
    query: {
      fetchPolicy: 'network-only',
    },
  },
});