import { reactive } from 'vue';
import {
  isLoggedIn,
  fetchMe,
  sendCode,
  loginWithCode,
  loginWithPassword,
  logout as apiLogout,
  type AuthUser,
} from '../api';

// Simple reactive auth store (no Pinia dependency)
export const authState = reactive({
  user: null as { id: number; email: string } | null,
  loading: false,
  error: '',

  get loggedIn() {
    return !!this.user;
  },

  async init() {
    if (!isLoggedIn()) return;
    try {
      const me = await fetchMe();
      this.user = { id: me.id, email: me.email };
    } catch {
      this.user = null;
    }
  },

  async login(email: string, codeOrPassword: string, mode: 'code' | 'password') {
    this.error = '';
    this.loading = true;
    try {
      const u = mode === 'code'
        ? await loginWithCode(email, codeOrPassword)
        : await loginWithPassword(email, codeOrPassword);
      this.user = { id: u.id, email: u.email };
      return u;
    } catch (e) {
      this.error = (e as Error).message;
      throw e;
    } finally {
      this.loading = false;
    }
  },

  async logout() {
    await apiLogout();
    this.user = null;
  },

  async sendCode(email: string) {
    await sendCode(email);
  },

  clearError() {
    this.error = '';
  },
});
