<template>
	<!-- ========== Left Sidebar Start ========== -->
	<div class="vertical-menu">

		<div class="h-100">

			<!--- Sidemenu -->
			<div id="sidebar-menu">
				<!-- Left Menu Start -->
				<ul class="metismenu list-unstyled">
					<li class="menu-title">Menu</li>
					<li :class="{'mm-active': routeName === 'jobs'}">
						<router-link :to="{name: 'jobs'}" class="waves-effect">
							<i class="bx bx-stats"></i>
							<span>Tiến trình</span>
						</router-link>
					</li>
					<li :class="{'mm-active': routeName === 'apps'}">
						<router-link :to="{name: 'apps'}" class="waves-effect">
							<i class="bx bx-tv"></i>
							<span>Ứng dụng</span>
						</router-link>
					</li>
				</ul>
			</div>
			<!-- Sidebar -->
		</div>
	</div>
	<!-- Left Sidebar End -->
</template>

<script>
	export default {
		name: 'SideBar',
		data: function () {
			return {
				channels: []
			}
		},
		computed: {
			routeName() {
				return this.$route.name
			}
		},
		methods: {
			expandMenu(event) {
				if (event.currentTarget && event.currentTarget.parentElement && event.currentTarget.nextElementSibling) {
					if (event.currentTarget.parentElement.classList.contains('mm-active')) {
						event.currentTarget.parentElement.classList.remove('mm-active')
						event.currentTarget.nextElementSibling.classList.remove('mm-show')
					} else {
						event.currentTarget.parentElement.classList.add('mm-active')
						event.currentTarget.nextElementSibling.classList.add('mm-show')
					}
				}
			}
		},
		created() {
		}
	}
</script>
