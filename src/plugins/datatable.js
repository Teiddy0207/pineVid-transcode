import DataTable from 'datatables.net-vue3'
import DataTableBs5 from 'datatables.net-bs5'
import vi from 'datatables.net-plugins/i18n/vi.json'

DataTable.use(DataTableBs5)

const baseConfig = {
	dom: '<\'row\'<\'col-sm-12\'tr>><\'row\'<\'col-sm-12 col-md-5\'i><\'col-sm-12 col-md-7\'p>>',
	language: vi,
	ordering: false,
	serverSide: true,
	stateSave: true,
	stateLoadCallback: function(settings) {
		const item = JSON.parse( localStorage.getItem( 'DataTables_DataTables_Table_0_/' + settings.ajax.url.substring(settings.ajax.url.lastIndexOf('/')+1)))
		return JSON.parse( localStorage.getItem('DataTables_DataTables_Table_0_'+location.pathname))
	},
	stateSaveCallback: function(settings,data) {
		const urlParams = new URLSearchParams(window.location.search)
		data.start = +urlParams.get('page') * +urlParams.get('limit')
		localStorage.setItem( 'DataTables_DataTables_Table_0_/' + settings.ajax.url.substring(settings.ajax.url.lastIndexOf('/')+1), JSON.stringify(data))
	},
}

export {baseConfig, DataTable}
