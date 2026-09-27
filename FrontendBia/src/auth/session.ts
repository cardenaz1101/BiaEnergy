import type { User } from '../api/types';

const TOKEN_KEY = 'bia.token';
const USER_KEY = 'bia.user';

function read(key: string): string | null {
  try {
    return localStorage.getItem(key);
  } catch {
    return null;
  }
}

function write(key: string, value: string | null) {
  try {
    if (value === null) localStorage.removeItem(key);
    else localStorage.setItem(key, value);
  } catch {
    return;
  }
}

export const session = {
  getToken: () => read(TOKEN_KEY),
  getUser: (): User | null => {
    try {
      return JSON.parse(read(USER_KEY) ?? 'null') as User | null;
    } catch {
      return null;
    }
  },
  save: (token: string, user: User) => {
    write(TOKEN_KEY, token);
    write(USER_KEY, JSON.stringify(user));
  },
  clear: () => {
    write(TOKEN_KEY, null);
    write(USER_KEY, null);
  },
};
