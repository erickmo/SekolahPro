import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useReducer,
} from 'react';
import AsyncStorage from '@react-native-async-storage/async-storage';
import api, { STORAGE_KEYS, post } from './api';
import type {
  AuthState,
  AuthTokens,
  LoginCredentials,
  School,
  User,
} from '../types';

// ─── Types ────────────────────────────────────────────────────────────────────

interface LoginResponse {
  user: User;
  school: School;
  tokens: AuthTokens;
}

type AuthAction =
  | { type: 'SET_LOADING'; payload: boolean }
  | { type: 'SET_AUTH'; payload: { user: User; school: School; tokens: AuthTokens } }
  | { type: 'CLEAR_AUTH' }
  | { type: 'UPDATE_USER'; payload: Partial<User> };

interface AuthContextValue extends AuthState {
  login: (credentials: LoginCredentials) => Promise<void>;
  logout: () => Promise<void>;
  refreshUser: () => Promise<void>;
}

// ─── Reducer ─────────────────────────────────────────────────────────────────

const initialState: AuthState = {
  user: null,
  school: null,
  tokens: null,
  isAuthenticated: false,
  isLoading: true,
};

function authReducer(state: AuthState, action: AuthAction): AuthState {
  switch (action.type) {
    case 'SET_LOADING':
      return { ...state, isLoading: action.payload };

    case 'SET_AUTH':
      return {
        ...state,
        user: action.payload.user,
        school: action.payload.school,
        tokens: action.payload.tokens,
        isAuthenticated: true,
        isLoading: false,
      };

    case 'CLEAR_AUTH':
      return { ...initialState, isLoading: false };

    case 'UPDATE_USER':
      return {
        ...state,
        user: state.user ? { ...state.user, ...action.payload } : null,
      };

    default:
      return state;
  }
}

// ─── Context ──────────────────────────────────────────────────────────────────

const AuthContext = createContext<AuthContextValue | null>(null);

// ─── Provider ─────────────────────────────────────────────────────────────────

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [state, dispatch] = useReducer(authReducer, initialState);

  // Rehydrate auth state from storage on mount
  useEffect(() => {
    async function rehydrate() {
      try {
        const [tokenRaw, userRaw, schoolRaw] = await AsyncStorage.multiGet([
          STORAGE_KEYS.ACCESS_TOKEN,
          STORAGE_KEYS.USER,
          STORAGE_KEYS.SCHOOL,
        ]);

        const token = tokenRaw[1];
        const user: User | null = userRaw[1] ? JSON.parse(userRaw[1]) : null;
        const school: School | null = schoolRaw[1] ? JSON.parse(schoolRaw[1]) : null;
        const refreshToken = await AsyncStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN);

        if (token && user && school) {
          dispatch({
            type: 'SET_AUTH',
            payload: {
              user,
              school,
              tokens: {
                accessToken: token,
                refreshToken: refreshToken ?? '',
                expiresIn: 0,
              },
            },
          });
        } else {
          dispatch({ type: 'SET_LOADING', payload: false });
        }
      } catch {
        dispatch({ type: 'SET_LOADING', payload: false });
      }
    }

    rehydrate();
  }, []);

  const login = useCallback(async (credentials: LoginCredentials) => {
    dispatch({ type: 'SET_LOADING', payload: true });
    try {
      const response = await post<LoginResponse>('/auth/login', credentials);
      const { user, school, tokens } = response;

      // Persist to AsyncStorage
      await AsyncStorage.multiSet([
        [STORAGE_KEYS.ACCESS_TOKEN, tokens.accessToken],
        [STORAGE_KEYS.REFRESH_TOKEN, tokens.refreshToken],
        [STORAGE_KEYS.USER, JSON.stringify(user)],
        [STORAGE_KEYS.SCHOOL, JSON.stringify(school)],
      ]);

      dispatch({ type: 'SET_AUTH', payload: { user, school, tokens } });
    } catch (error) {
      dispatch({ type: 'SET_LOADING', payload: false });
      throw error;
    }
  }, []);

  const logout = useCallback(async () => {
    try {
      // Notify backend (best-effort; ignore failures)
      await api.post('/auth/logout').catch(() => undefined);
    } finally {
      await AsyncStorage.multiRemove([
        STORAGE_KEYS.ACCESS_TOKEN,
        STORAGE_KEYS.REFRESH_TOKEN,
        STORAGE_KEYS.USER,
        STORAGE_KEYS.SCHOOL,
      ]);
      dispatch({ type: 'CLEAR_AUTH' });
    }
  }, []);

  const refreshUser = useCallback(async () => {
    try {
      const response = await post<{ user: User }>('/auth/me');
      dispatch({ type: 'UPDATE_USER', payload: response.user });
      await AsyncStorage.setItem(STORAGE_KEYS.USER, JSON.stringify(response.user));
    } catch {
      // Silently ignore — user stays as-is
    }
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({ ...state, login, logout, refreshUser }),
    [state, login, logout, refreshUser],
  );

  return React.createElement(AuthContext.Provider, { value }, children);
}

// ─── Hook ─────────────────────────────────────────────────────────────────────

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

export { AuthContext };
