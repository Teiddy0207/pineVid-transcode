import { notify } from '@kyvg/vue3-notification'

export default {
	namespaced: true,
	state: () => ({
		type: '',
		message: ''
	}),
	getters: {},
	actions: {
		setError({ commit }, error) {
			commit('setError', error)
		},
		setMessage({ commit }, error) {
			commit('setMessage', error)
		}
	},
	mutations: {
		setError(state, error) {
			if (error) {
				if (
					Object.prototype.hasOwnProperty.call(error, 'data') &&
					Object.prototype.hasOwnProperty.call(error.data, 'message') &&
					Object.prototype.hasOwnProperty.call(error, 'status')
				) {
					switch (error.status) {
						case 200:
							state.type = 'success'
							break
						case 201:
							state.type = 'success'
							break
						case 400:
							state.type = 'warn'
							break
						case 401:
							state.type = 'warn'
							break
						case 404:
							state.type = 'warn'
							break
						case 500:
							state.type = 'error'
							break
						default:
							state.type = 'error'
					}
					state.message = error.data.message
				} else if (Object.prototype.hasOwnProperty.call(error, 'status')) {
					switch (error.status) {
						case 400:
							state.type = 'warn'
							state.message = 'Truy vấn không hợp lệ'
							break
						case 401:
							state.type = 'warn'
							state.message = 'Chưa đăng nhập'
							break
						case 404:
							state.type = 'warn'
							state.message = 'Không tìm thấy dữ liệu'
							break
						case 500:
							state.type = 'error'
							state.message = 'Lỗi hệ thống'
							break
						default:
							state.type = 'error'
							state.message = 'Lỗi hệ thống'
					}
				} else {
					state.type = 'error'
					state.message = 'Lỗi hệ thống'
				}
				notify({
					title: 'Thông báo',
					text: state.message,
					type: state.type,
					duration: 7000,
					speed: 1000,
					position: 'top right'
				})
			}
		},
		setMessage(state, options) {
			notify({
				title: 'Thông báo',
				text: options.message,
				type: options.type,
				duration: 5000,
				speed: 1000,
				position: 'top right'
			})
		}
	}
}
