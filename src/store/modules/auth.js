import auth from '@/api/auth'
import admin from '@/api/admin'

export default {
    namespaced: true,

    state: () => ({
        user: null
    }),
    getters: {
        getUser(state) {
            return state.user
        },
    },
    actions: {
        async login({dispatch, commit}, {username, password, remember}) {
            let response = await auth.login(username, password, remember)
            if (response.status === 200) {
                dispatch('profile')
            } else {
                dispatch('message/setError', {status: response.status, data: response.data}, {root: true})
            }
        },
        async logout({dispatch, commit}) {
            let response = await auth.logout()
            if (response.status === 200) {
                commit('clearUser')
            } else {
                dispatch('message/setError', {status: response.status, data: response.data})
            }
        },
        async profile({dispatch, commit}) {
            let response = await admin.profile()
            if (response.status === 200) {
                commit('setUser', response.data)
            } else if (response.status !== 401) {
                dispatch('message/setError', {status: response.status, data: response.data}, {root: true})
            }
        },
        clearUser({dispatch, commit}) {
            commit('clearUser')
        }
    },
    mutations: {
        setUser(state, user) {
            state.user = user
        },
        clearUser(state) {
            state.user = null
        }
    }
}
