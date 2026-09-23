<template>
	<header id="page-topbar">
		<div class="navbar-header">
			<div class="d-flex">
				<!-- LOGO -->
				<div class="navbar-brand-box">
					<a href="index.html" class="logo logo-dark">
						<span class="logo-sm">
							<img src="@/assets/logo.png" alt="" height="22" />
						</span>
						<span class="logo-lg">
							<img src="@/assets/logo.png" alt="" height="17" />
						</span>
					</a>
					<a href="index.html" class="logo logo-light">
						<span class="logo-sm">
							<img src="@/assets/logo.png" alt="" height="22" />
						</span>
						<span class="logo-lg">
							<img src="@/assets/logo.png" alt="" height="68" />
						</span>
					</a>
				</div>
				<button type="button" class="btn btn-sm px-3 font-size-16 header-item waves-effect" @click="changeSideBarSize">
					<i class="fa fa-fw fa-bars"></i>
				</button>
			</div>

			<div class="d-flex">
				<div class="dropdown d-inline-block">
					<button type="button" class="btn header-item waves-effect" data-bs-toggle="dropdown" aria-haspopup="true"
									aria-expanded="false">
						<div class="avatar-sm mx-auto mb-4">
							<span class="avatar-title rounded-circle bg-primary bg-soft text-primary header-profile-user font-size-16">A</span>
						</div>
						<span class="d-none d-xl-inline-block ms-1">{{ getUser.full_name }}</span>
						<i class="mdi mdi-chevron-down d-none d-xl-inline-block"></i>
					</button>
					<div class="dropdown-menu dropdown-menu-end">
						<!-- item-->
						<router-link :to="{name: 'profile'}" class="dropdown-item">
							<i class="bx bx-user font-size-16 align-middle me-1"></i>
							<span>Tài khoản</span>
						</router-link>
						<div class="dropdown-divider"></div>
						<a class="dropdown-item text-danger" href="javascript:void(0);" @click="logout">
							<i class="bx bx-power-off font-size-16 align-middle me-1 text-danger"></i>
							<span>Đăng xuất</span>
						</a>
					</div>
				</div>
			</div>
		</div>
	</header>
</template>

<script>
	import {mapGetters} from 'vuex'

	export default {
		name: 'TopBar',
		computed: {
			...mapGetters('auth', ['getUser'])
		},
		methods: {
			changeSideBarSize() {
				if (window.innerWidth < 993) {
					if (document.body.classList.contains('sidebar-enable')) {
						document.body.classList.remove('sidebar-enable')
					} else {
						document.body.classList.add('sidebar-enable')
					}
					if (document.body.classList.contains('vertical-collpsed')) {
						document.body.classList.remove('vertical-collpsed')
					}
				} else {
					if (
						document.body.classList.contains('sidebar-enable') &&
						document.body.classList.contains('vertical-collpsed')
					) {
						document.body.classList.remove('sidebar-enable')
						document.body.classList.remove('vertical-collpsed')
					} else {
						document.body.classList.add('sidebar-enable')
						document.body.classList.add('vertical-collpsed')
					}
				}
			},
			fullscreen() {
				if (
					!document.fullscreenElement &&
					!document.mozFullScreenElement &&
					!document.webkitFullscreenElement
				) {
					if (document.documentElement.requestFullscreen) {
						document.documentElement.requestFullscreen()
					} else if (document.documentElement.mozRequestFullScreen) {
						document.documentElement.mozRequestFullScreen()
					} else if (document.documentElement.webkitRequestFullscreen) {
						document.documentElement.webkitRequestFullscreen(
							Element.ALLOW_KEYBOARD_INPUT
						)
					}
				} else {
					if (document.cancelFullScreen) {
						document.cancelFullScreen()
					} else if (document.mozCancelFullScreen) {
						document.mozCancelFullScreen()
					} else if (document.webkitCancelFullScreen) {
						document.webkitCancelFullScreen()
					}
				}
			},
			async logout() {
				try {
					await this.$store.dispatch('auth/logout')
					this.$router.push('/login')
				} catch (error) {
					console.log(error)
				}
			}
		}
	}
</script>
