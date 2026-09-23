import dashboard from '@/api/dashboard'

export default {
	namespaced: true,
	state: () => ({
		dashboard: null
	}),
	getters: {
		getDashboard(state) {
			return state.dashboard
		}
	},
	actions: {
		async loadDashboard({dispatch, commit}) {
			let response = await dashboard.dashboard()
			if (response.status === 200) {
				commit('setDashboard', response.data)
			} else {
				dispatch('error/setError', {
					status: response.status,
					data: response.data
				})
			}
		}
	},
	mutations: {
		setDashboard(state, dashboard) {
			state.dashboard = dashboard
		}
	}
}
