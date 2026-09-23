import axios from 'axios'
import store from '@/store'

export default axios.create({
	baseURL: '/api/backend',
	validateStatus: (status) => {
		if (status === 401) {
			store.dispatch('auth/clearUser')
		}
		return true
	},
})
