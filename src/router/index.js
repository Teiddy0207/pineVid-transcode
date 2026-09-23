import {createRouter, createWebHistory} from 'vue-router'
import store from '@/store'

const routes = [
	{
		path: '/',
		name: 'home',
		component: () => import('@/layouts/DefaultLayout.vue'),
		meta: {
			layout: 'DefaultLayout',
			requiresAuth: true
		},
		redirect: '/jobs',
		children: [
			{
				name: 'jobs',
				path: '/',
				component: () => import('@/views/job/Index.vue')
			},
			{
				name: 'apps',
				path: '/apps',
				component: () => import('@/views/app/Index.vue')
			},
			{
				name: 'profile',
				path: '/profile',
				component: () => import('@/views/profile/Index.vue')
			},
		]
	},
	{
		path: '/login',
		name: 'login',
		component: () => import('@/views/auth/Login.vue'),
		meta: {
			layout: 'AuthLayout',
			requiresAuth: false
		}
	},
	{path: '/:path(.*)', component: () => import('@/views/auth/Login.vue')}
]

const router = createRouter({
	history: createWebHistory(import.meta.env.BASE_URL),
	routes,
	linkActiveClass: '',
	linkExactPathActiveClass: '',
	linkExactActiveClass: '',
	scrollBehavior() {
		return {x: 0, y: 0}
	}
})

router.beforeEach(async (to, from, next) => {
	const user = store.getters['auth/getUser']
	if (to.matched.some((record) => record.meta.requiresAuth) && user === null) {
		try {
			await store.dispatch('auth/profile')
			const subUser = store.getters['auth/getUser']
			if (subUser === null) {
				next('/login')
			} else {
				next()
			}
		} catch (error) {
			next('/login')
		}
	} else if (user !== null) {
		switch (to.name) {
			case 'login':
				next({path: '/'})
				break
			default:
				next()
				break
		}
	} else {
		next()
	}
})

router.afterEach((to, from) => {
	let tooltipList = [].slice.call(document.querySelectorAll('[role="tooltip"]'))
	tooltipList.map(function (el) {
		el.remove()
	})
})

export default router
