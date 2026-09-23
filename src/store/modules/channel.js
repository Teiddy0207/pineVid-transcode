import channel from '@/api/channel'

export default {
	namespaced: true,
	state: () => ({
		channels: []
	}),
	getters: {
		getChannels(state) {
			return state.channels
		}
	},
	actions: {
		async loadChannels({dispatch, commit}) {
			let response = await channel.index({params: {page: -1}})
			if (response.status === 200) {
				commit('setChannels', response.data.data)
			} else {
				dispatch('error/setError', {
					status: response.status,
					data: response.data
				})
			}
		}
	},
	mutations: {
		setChannels(state, channels) {
			state.channels = channels
		}
	}
}
