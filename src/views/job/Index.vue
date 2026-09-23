<template>
	<div class="row">
		<div class="col-12">
			<div class="page-title-box d-sm-flex align-items-center justify-content-between">
				<h4 class="mb-sm-0 font-size-18">Quản lý tiến trình</h4>
				<div class="page-title-right">
					<a href="javascript:void(0)" @click="elements.jobModalCreate.show()"
						 class="btn btn-success waves-effect btn-label waves-light">
						<i class="bx bx-plus label-icon"></i>
						Thêm mới
					</a>
				</div>
			</div>
		</div>
		<div class="col-lg-12">
			<div class="card">
				<div class="card-body">
					<div class="row mb-3">
						<div class="col-12 col-lg-3">
							<label>Tìm kiếm</label>
							<input type="text" class="form-control" v-model="query.search" @change="forceRender" placeholder="Tìm kiếm ..."/>
						</div>
						<div class="col-12 col-lg-3">
							<label>Ứng dụng</label>
							<Multiselect v-model="query.app" :options="options.apps" noOptionsText="Không có lựa chọn"
													 noResultsText="Không có kết quả" :searchable="true" :canDeselect="true" :canClear="true"
													 @select="forceRender" @clear="forceRender" @search-change="searchApp"/>
						</div>
						<div class="col-12 col-lg-3">
							<label>Trạng thái</label>
							<Multiselect v-model="query.status" :options="options.status" noOptionsText="Không có lựa chọn"
													 noResultsText="Không có kết quả" :searchable="false" :canDeselect="true" :canClear="true"
													 @select="forceRender" @clear="forceRender"/>
						</div>
					</div>

					<data-table class="table table-bordered dt-responsive nowrap w-100" :options="dataTableConfigs"
											ref="jobTable">
					</data-table>
				</div>
			</div>
		</div>
	</div>
	<div class="modal fade" tabindex="-1" role="dialog" aria-labelledby="job-modal-create-label" aria-hidden="true"
			 data-bs-backdrop="static" id="job-modal-create">
		<div class="modal-dialog">
			<div class="modal-content">
				<div class="modal-header">
					<h5 class="modal-title" id="job-modal-create-label">Thêm tiến trình</h5>
				</div>
				<div class="modal-body">
					<form @submit.prevent="submitCreate" method="post">
						<div class="row">
							<div class="col-12">
								<div class="mb-3">
									<label class="form-label required">Ứng dụng</label>
									<Multiselect v-model="app" :options="options.apps" noOptionsText="Không có lựa chọn"
															 noResultsText="Không có kết quả" :searchable="true" :canDeselect="false" :canClear="false"
															 :object="true" @search-change="searchApp"/>
									<ul class="parsley-errors-list filled" v-if="errors.app">
										<li class="parsley-required">{{ errors.app }}</li>
									</ul>
								</div>
							</div>
							<div class="col-12">
								<div class="mb-3">
									<label class="form-label required">ID yêu cầu</label>
									<div class="input-group">
										<input type="text" class="form-control" placeholder="ID yêu cầu" aria-label="ID yêu cầu"
													 aria-describedby="request-addon" v-model="requestID" @focus="errors.requestID = false">
										<button class="btn btn-success" type="button" id="request-addon" title="Tạo ngẫu nhiên"
														@click="randomRequestID">
											<i class="mdi mdi-shuffle"></i>
										</button>
									</div>
									<ul class="parsley-errors-list filled" v-if="errors.requestID">
										<li class="parsley-required">{{ errors.requestID }}</li>
									</ul>
								</div>
							</div>
							<div class="col-12">
								<div class="mb-3">
									<label class="form-label required">Tập tin nguồn</label>
									<input type="text" class="form-control" v-model="source" @focus="errors.source = false">
									<ul class="parsley-errors-list filled" v-if="errors.source">
										<li class="parsley-required">{{ errors.source }}</li>
									</ul>
								</div>
							</div>
							<div class="col-12">
								<div class="mb-3">
									<label class="form-label required">Thư mục lưu</label>
									<input type="text" class="form-control" v-model="destination" @focus="errors.destination = false">
									<ul class="parsley-errors-list filled" v-if="errors.destination">
										<li class="parsley-required">{{ errors.destination }}</li>
									</ul>
								</div>
							</div>
							<div class="col-4">
								<div class="mb-3">
									<div class="form-check">
										<input class="form-check-input" type="checkbox" id="input-hls" v-model="hls">
										<label class="form-check-label" for="input-hls">HLS</label>
									</div>
								</div>
							</div>
							<div class="col-4">
								<div class="mb-3">
									<div class="form-check">
										<input class="form-check-input" type="checkbox" id="input-dash" v-model="dash">
										<label class="form-check-label" for="input-dash">DASH</label>
									</div>
								</div>
							</div>
							<div class="col-4">
								<div class="mb-3">
									<div class="form-check">
										<input class="form-check-input" type="checkbox" id="input-encrypt" v-model="encrypt">
										<label class="form-check-label" for="input-encrypt">Mã hóa</label>
									</div>
								</div>
							</div>
							<div class="col-12" v-if="encrypt">
								<div class="mb-3">
									<label class="form-label required">Key mã hóa</label>
									<div class="input-group">
										<input type="text" class="form-control" placeholder="Key mã hóa" aria-label="Key mã hóa"
													 aria-describedby="key-addon" v-model="key" @focus="errors.key = false">
										<button class="btn btn-success" type="button" id="key-addon" title="Tạo ngẫu nhiên"
														@click="randomKey">
											<i class="mdi mdi-shuffle"></i>
										</button>
									</div>
									<ul class="parsley-errors-list filled" v-if="errors.key">
										<li class="parsley-required">{{ errors.key }}</li>
									</ul>
								</div>
							</div>
							<div class="col-12" v-if="encrypt">
								<div class="mb-3">
									<label class="form-label required">URL trả key</label>
									<input type="text" class="form-control" v-model="keyUri" @focus="errors.keyUri = false">
									<ul class="parsley-errors-list filled" v-if="errors.keyUri">
										<li class="parsley-required">{{ errors.keyUri }}</li>
									</ul>
								</div>
							</div>
							<div class="col-12">
								<div class="mb-3">
									<label class="form-label">Webhook</label>
									<input type="text" class="form-control" v-model="callback" @focus="errors.callback = false">
									<ul class="parsley-errors-list filled" v-if="errors.callback">
										<li class="parsley-required">{{ errors.callback }}</li>
									</ul>
								</div>
							</div>
							<div class="col-12">
								<div class="mb-3">
									<label class="form-label required">Trạng thái</label>
									<Multiselect v-model="status" :options="options.status" noOptionsText="Không có lựa chọn"
															 noResultsText="Không có kết quả" :searchable="false" :canDeselect="false" :canClear="false"/>
									<ul class="parsley-errors-list filled" v-if="errors.status">
										<li class="parsley-required">{{ errors.status }}</li>
									</ul>
								</div>
							</div>
						</div>
						<div class="button-items">
							<button type="submit" class="btn btn-success waves-effect btn-label waves-light w-md"><i
											class="bx bx-save label-icon"></i>Lưu
							</button>
							<button type="button" class="btn btn-light waves-effect waves-light w-md"
											@click="closeModal">Đóng
							</button>
						</div>
					</form>
				</div>
			</div>
		</div>
	</div>
	<div class="modal fade" tabindex="-1" role="dialog" aria-labelledby="job-modal-update-label" aria-hidden="true"
			 data-bs-backdrop="static" id="job-modal-update">
		<div class="modal-dialog">
			<div class="modal-content">
				<div class="modal-header">
					<h5 class="modal-title" id="job-modal-update-label">Cập nhật tiến trình</h5>
				</div>
				<div class="modal-body">
					<div class="spinner-border text-primary m-1" role="status" v-if="isModalLoading">
						<span class="sr-only">Đang tải ...</span>
					</div>
					<form @submit.prevent="submitUpdate" method="post" v-else>
						<div class="row">
							<div class="col-12">
								<div class="mb-3">
									<label class="form-label required">Ứng dụng</label>
									<Multiselect v-model="app" :options="options.apps" noOptionsText="Không có lựa chọn"
															 noResultsText="Không có kết quả" :searchable="true" :canDeselect="false" :canClear="false"
															 :object="true" @search-change="searchApp"/>
									<ul class="parsley-errors-list filled" v-if="errors.app">
										<li class="parsley-required">{{ errors.app }}</li>
									</ul>
								</div>
							</div>
							<div class="col-12">
								<div class="mb-3">
									<label class="form-label required">ID yêu cầu</label>
									<div class="input-group">
										<input type="text" class="form-control" placeholder="ID yêu cầu" aria-label="ID yêu cầu"
													 aria-describedby="request-addon-update" v-model="requestID" @focus="errors.requestID = false">
										<button class="btn btn-success" type="button" id="request-addon-update" title="Tạo ngẫu nhiên"
														@click="randomRequestID">
											<i class="mdi mdi-shuffle"></i>
										</button>
									</div>
									<ul class="parsley-errors-list filled" v-if="errors.requestID">
										<li class="parsley-required">{{ errors.requestID }}</li>
									</ul>
								</div>
							</div>
							<div class="col-12">
								<div class="mb-3">
									<label class="form-label required">Tập tin nguồn</label>
									<input type="text" class="form-control" v-model="source" @focus="errors.source = false">
									<ul class="parsley-errors-list filled" v-if="errors.source">
										<li class="parsley-required">{{ errors.source }}</li>
									</ul>
								</div>
							</div>
							<div class="col-12">
								<div class="mb-3">
									<label class="form-label required">Thư mục lưu</label>
									<input type="text" class="form-control" v-model="destination" @focus="errors.destination = false">
									<ul class="parsley-errors-list filled" v-if="errors.destination">
										<li class="parsley-required">{{ errors.destination }}</li>
									</ul>
								</div>
							</div>
							<div class="col-4">
								<div class="mb-3">
									<div class="form-check">
										<input class="form-check-input" type="checkbox" id="input-hls-update" v-model="hls">
										<label class="form-check-label" for="input-hls-update">HLS</label>
									</div>
								</div>
							</div>
							<div class="col-4">
								<div class="mb-3">
									<div class="form-check">
										<input class="form-check-input" type="checkbox" id="input-dash-update" v-model="dash">
										<label class="form-check-label" for="input-dash-update">DASH</label>
									</div>
								</div>
							</div>
							<div class="col-4">
								<div class="mb-3">
									<div class="form-check">
										<input class="form-check-input" type="checkbox" id="input-encrypt-update" v-model="encrypt">
										<label class="form-check-label" for="input-encrypt-update">Mã hóa</label>
									</div>
								</div>
							</div>
							<div class="col-12" v-if="encrypt">
								<div class="mb-3">
									<label class="form-label required">Key mã hóa</label>
									<div class="input-group">
										<input type="text" class="form-control" placeholder="Key mã hóa" aria-label="Key mã hóa"
													 aria-describedby="key-addon-update" v-model="key" @focus="errors.key = false">
										<button class="btn btn-success" type="button" id="key-addon-update" title="Tạo ngẫu nhiên"
														@click="randomKey">
											<i class="mdi mdi-shuffle"></i>
										</button>
									</div>
									<ul class="parsley-errors-list filled" v-if="errors.key">
										<li class="parsley-required">{{ errors.key }}</li>
									</ul>
								</div>
							</div>
							<div class="col-12" v-if="encrypt">
								<div class="mb-3">
									<label class="form-label required">URL trả key</label>
									<input type="text" class="form-control" v-model="keyUri" @focus="errors.keyUri = false">
									<ul class="parsley-errors-list filled" v-if="errors.keyUri">
										<li class="parsley-required">{{ errors.keyUri }}</li>
									</ul>
								</div>
							</div>
							<div class="col-12">
								<div class="mb-3">
									<label class="form-label">Webhook</label>
									<input type="text" class="form-control" v-model="callback" @focus="errors.callback = false">
									<ul class="parsley-errors-list filled" v-if="errors.callback">
										<li class="parsley-required">{{ errors.callback }}</li>
									</ul>
								</div>
							</div>
							<div class="col-12">
								<div class="mb-3">
									<label class="form-label required">Trạng thái</label>
									<Multiselect v-model="status" :options="options.status" noOptionsText="Không có lựa chọn"
															 noResultsText="Không có kết quả" :searchable="false" :canDeselect="false" :canClear="false"/>
									<ul class="parsley-errors-list filled" v-if="errors.status">
										<li class="parsley-required">{{ errors.status }}</li>
									</ul>
								</div>
							</div>
						</div>
						<div class="button-items">
							<button type="submit" class="btn btn-success waves-effect btn-label waves-light w-md"><i
											class="bx bx-save label-icon"></i>Lưu
							</button>
							<button type="button" class="btn btn-light waves-effect waves-light w-md"
											@click="closeModal">Đóng
							</button>
						</div>
					</form>
				</div>
			</div>
		</div>
	</div>
</template>

<script>
	import {Modal} from 'bootstrap'
	import job from '@/api/job'
	import app from '@/api/app'
	import * as yup from 'yup'
	import {useForm} from 'vee-validate'
	import Swal from 'sweetalert2'
	import {mapGetters} from 'vuex'
	import Multiselect from '@vueform/multiselect'
	import {baseConfig, DataTable} from '../../plugins/datatable'
	import {h, render} from 'vue'
import moment from 'moment'

	export default {
		name: 'JobIndex',
		components: {
			DataTable,
			Multiselect
		},
		setup() {
			const simpleSchema = yup.object({
				app: yup.object().typeError('Dữ liệu không được bỏ trống').required(),
				requestID: yup.string().required(),
				source: yup.string().required(),
				destination: yup.string().required(),
				key: yup.string().when('encrypt', {
					is: (encrypt) => encrypt === true,
					then: yup.string().required(),
					otherwise: yup.string().notRequired()
				}),
				keyUri: yup.string().when('encrypt', {
					is: (encrypt) => encrypt === true,
					then: yup.string().required(),
					otherwise: yup.string().notRequired()
				}),
			})
			const {useFieldModel, validate, validateField} = useForm({
				initialValues: {
					app: null,
					requestID: '',
					source: '',
					destination: '',
					hls: false,
					dash: false,
					encrypt: false,
					key: '',
					keyUri: '',
					callback: '',
					status: 'waiting',
				},
				validationSchema: simpleSchema,
				validateOnMount: false
			})
			const [app, requestID, source, destination, hls, dash, encrypt, key, keyUri, callback, status] = useFieldModel(['app', 'requestID', 'source', 'destination', 'hls', 'dash', 'encrypt', 'key', 'keyUri', 'callback', 'status'])
			return {
				validate,
				validateField,
				app,
				requestID,
				source,
				destination,
				hls,
				dash,
				encrypt,
				key,
				keyUri,
				callback,
				status
			}
		},
		data: function () {
			return {
				dataTableConfigs: {
					...baseConfig,
					... {
						ajax: {
							url: '/api/backend/jobs',
							data: this.tableQuery
						},
						columns: [
							{
								className: 'align-middle',
								data: 'source',
								name: 'source',
								orderable: false,
								title: 'Tiến trình',
								render: function (data, type, row, meta) {
									let source = `<p style="white-space:break-spaces;">Tập tin nguồn: ${row.source}</p>`
									let destination = `<p style="white-space:break-spaces;">Thư mục đích: ${row.destination}</p>`
									let app = `<span class="badge rounded-pill bg-primary font-size-12 me-2">${row.app ? row.app.name : '????'}</span>`
									let tmp = row.hls ? `<span class="badge rounded-pill bg-success font-size-12 me-2">HLS</span>` : ''
									tmp += row.dash ? `<span class="badge rounded-pill bg-warning font-size-12 me-2">DASH</span>` : ''
									tmp += row.encrypt ? `<span class="badge rounded-pill bg-danger font-size-12">Mã hóa</span>` : ''
									return `${source}${destination}<p class="mb-0">${app}${tmp}</p>`
								}
							},
							{
								className: 'align-middle text-center',
								data: 'status',
								name: 'status',
								orderable: false,
								title: 'Trạng thái',
								width: '120px',
								render: function (data, type, row, meta) {
									switch (data) {
										case 'waiting':
											return '<span class="badge rounded-pill bg-primary font-size-12">Chờ</span>'
										case 'running':
											return `<span class="badge rounded-pill bg-warning font-size-12">Đang chạy: ${row.progress}%</span>`
										case 'done':
											return '<span class="badge rounded-pill bg-success font-size-12">Hoàn thành</span>'
										case 'error':
											return '<span class="badge rounded-pill bg-danger font-size-12">Lỗi</span>'
										default:
											return '<span class="badge rounded-pill bg-info font-size-12">????</span>'
									}
								}
							},
							{
								className: 'text-center align-middle',
								data: 'createdAt',
								name: 'id',
								title: 'Ngày tao',
								orderable: true,
								width: '120px',
								render: function (data) {
									return moment(data).format('DD/MM/YYYY HH:mm:ss')
								}
							},
							{
								className: 'text-center align-middle',
								data: 'id',
								title: 'Hành động',
								width: '120px',
								orderable: false,
								render: function () {
									return ''
								}
							}
						],
						ordering: true,
						order: [[2, 'desc']],
						serverSide: true,
						rowCallback: this.rowCallback
					}
				},
				elements: {
					jobTable: null,
					jobModalCreate: null,
					jobModalUpdate: null
				},
				errors: {},
				isSubmitting: false,
				isModalLoading: false,
				job: {},
				options: {
					apps: [],
					status: [{value: 'waiting', label: 'Chờ'}, {value: 'running', label: 'Đang chạy'}, {value: 'done', label: 'Hoàn thành'}, {value: 'error', label: 'Lỗi'}]
				},
				query: {
					app: '',
					search: '',
					status: '',
				},
			}
		},
		computed: {
			...mapGetters('auth', ['getUser'])
		},
		watch: {},
		methods: {
			tableQuery(d) {
				console.log(d)
				let order = ''
				let dir = 'asc'
				if (d.order && d.order.length > 0) {
					order = d.columns[d.order[0].column].name
					dir = d.order[0].dir
				}
				return {
					draw: d.draw,
					limit: d.length,
					page: d.start / d.length,
					search: this.query.search,
					status: this.query.status,
					appID: this.query.app,
					order: order,
					dir: dir
				}
			},
			forceRender() {
				this.$nextTick(function () {
					this.$refs.jobTable.dt().ajax.reload(null, false)
				})
			},
			resetForm() {
				this.app = null
				this.requestID = ''
				this.source = ''
				this.destination = ''
				this.hls = false
				this.dash = false
				this.encrypt = false
				this.key = ''
				this.keyUri = ''
				this.callback = ''
				this.status = 'waiting'
			},
			closeModal() {
				this.resetForm()
				this.elements.jobModalCreate?.hide()
				this.elements.jobModalUpdate?.hide()
			},
			rowCallback(row, data) {
				let group = h('div', {class: 'button-items'}, [])
				let update = h('button', {
						class: 'btn btn-outline-primary waves-effect waves-light btn-sm',
						onClick: () => {
							this.updateJob(data.id)
						}
					}, [h('i', {class: 'bx bx-pencil font-size-16'})]
				)
				let remove = h('button', {
						class: 'btn btn-outline-danger waves-effect waves-light btn-sm',
						onClick: () => {
							this.submitDelete(data.id)
						}
					}, [h('i', {class: 'bx bx-trash font-size-16'})]
				)
				group.children.push(update)
				group.children.push(remove)
				render(group, row.lastElementChild)
			},
			randomRequestID() {
				let result = ''
				let characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
				let charactersLength = characters.length
				for (let i = 0; i < 12; i++) {
					result += characters.charAt(Math.floor(Math.random() * charactersLength))
				}
				this.requestID = result
			},
			randomKey() {
				let result = ''
				let characters = 'ABCDEF0123456789'
				let charactersLength = characters.length
				for (let i = 0; i < 32; i++) {
					result += characters.charAt(Math.floor(Math.random() * charactersLength))
				}
				this.key = result
			},
			async searchApp(query) {
				let res = await app.search({
					params: {search: query}
				})
				if (res.status === 200) {
					this.options.apps = res.data.map((item) => {
						return {value: item.id, label: item.name}
					})
				} else {
					this.options.apps = []
				}
			},
			async updateJob(id) {
				this.job.id = id
				this.elements.jobModalUpdate?.show()
				this.isModalLoading = true
				this.errors = {}
				let res = await job.read(id)
				if (res.status === 200) {
					this.app = res.data.app ? {value: res.data.app.id, label: res.data.app.name} : null
					this.requestID = res.data.requestID
					this.source = res.data.source
					this.destination = res.data.destination
					this.hls = res.data.hls
					this.dash = res.data.dash
					this.encrypt = res.data.encrypt
					this.key = res.data.key
					this.keyUri = res.data.keyUri
					this.callback = res.data.callback
					this.status = res.data.status
				} else {
					this.elements.jobModalUpdate?.hide()
					await this.$store.dispatch('message/setError', res)
				}
				this.isModalLoading = false
			},
			async submitCreate() {
				this.isSubmitting = true
				const {valid, errors} = await this.validate()
				this.errors = errors
				if (valid) {
					let res = await job.create({
						appID: this.app.value,
						requestID: this.requestID,
						source: this.source,
						destination: this.destination,
						hls: this.hls,
						dash: this.dash,
						encrypt: this.encrypt,
						key: this.key,
						keyUri: this.keyUri,
						callback: this.callback,
						status: this.status,
					})
					await this.$store.dispatch('message/setError', res)
					if (res.status === 201) {
						this.elements.jobModalCreate?.hide()
						this.forceRender()
						this.resetForm()
					} else if (res.status === 422 && typeof res.data.errors === 'object') {
						this.errors = res.data.errors
					}
				}
				this.isSubmitting = false
			},
			async submitUpdate() {
				this.isSubmitting = true
				const {valid, errors} = await this.validate()
				this.errors = errors
				if (valid) {
					let res = await job.update(this.job.id, {
						appID: this.app.value,
						requestID: this.requestID,
						source: this.source,
						destination: this.destination,
						hls: this.hls,
						dash: this.dash,
						encrypt: this.encrypt,
						key: this.key,
						keyUri: this.keyUri,
						callback: this.callback,
						status: this.status,
					})
					await this.$store.dispatch('message/setError', res)
					if (res.status === 200) {
						this.elements.jobModalUpdate?.hide()
						this.forceRender()
						this.resetForm()
					} else if (res.status === 422 && typeof res.data.errors === 'object') {
						this.errors = res.data.errors
					}
				}
				this.isSubmitting = false
			},
			async submitDelete(id) {
				let confirm = await Swal.fire({
					title: 'Cảnh báo',
					text: 'Bạn có chắc muốn xóa?',
					icon: 'question',
					confirmButtonText: 'Có, chắc chắn',
					cancelButtonText: 'Không',
					confirmButtonColor: '#f46a6a',
					customClass: {
						closeButton: 'btn btn-primary waves-effect waves-light',
						confirmButton: 'btn btn-danger waves-effect waves-light',
						cancelButton: 'btn btn-primary waves-effect waves-light'
					},
					showConfirmButton: true,
					showCancelButton: true,
					focusConfirm: false,
					focusCancel: true,
					showLoaderOnConfirm: true,
					preConfirm: () => {
					}
				})
				if (confirm.isConfirmed) {
					let res = await job.delete(id)
					this.forceRender()
					await this.$store.dispatch('message/setError', res)
				}
			}
		},
		mounted() {
			this.searchApp()
			this.$nextTick(function () {
				this.elements.jobModalCreate = new Modal(document.getElementById('job-modal-create'))
				this.elements.jobModalUpdate = new Modal(document.getElementById('job-modal-update'))
			})
		},
		unmounted() {
		}
	}
</script>
