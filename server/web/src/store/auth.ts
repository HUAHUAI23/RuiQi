import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { AuthState, User } from '@/types/auth'

export const useAuthStore = create<AuthState>()(
    persist(
        (set) => ({
            user: null,
            token: null,
            isAuthenticated: false,
            needPasswordReset: false,

            login: (token: string, user: User) => {
                set({
                    user,
                    token,
                    isAuthenticated: true,
                    needPasswordReset: user.needReset || false,
                })
            },

            logout: () => {
                set({
                    user: null,
                    token: null,
                    isAuthenticated: false,
                    needPasswordReset: false,
                })
            },

            setUser: (user: User) => {
                set({
                    user,
                    needPasswordReset: user.needReset || false,
                })
            },
        }),
        {
            name: 'auth-storage',
            // Only persist these fields
            partialize: (state) => ({
                token: state.token,
                user: state.user,
                isAuthenticated: state.isAuthenticated,
                needPasswordReset: state.needPasswordReset,
            }),
        }
    )
)

export default useAuthStore 