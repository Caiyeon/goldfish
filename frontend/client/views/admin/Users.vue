<template>
  <div>

    <div class="tile is-ancestor">
      <div class="tile is-parent is-vertical">
        <article class="tile is-child box">

          <!-- Tab navigation -->
          <div class="tabs is-medium is-boxed is-fullwidth">
            <ul>
              <li v-bind:class="tabName === 'token' ? 'is-active' : ''" v-on:click="switchTab(0)"><a>Tokens</a></li>
              <li v-bind:class="tabName === 'userpass' ? 'is-active' : ''" v-on:click="switchTab(1)"><a>Userpass</a></li>
              <li v-bind:class="tabName === 'approle' ? 'is-active' : ''" v-on:click="switchTab(2)"><a>Approle</a></li>
              <li class="is-disabled"><a>Certificates</a></li>
            </ul>
          </div>

          <!-- Tokens tab -->
          <div v-if="tabName === 'token'" class="tile is-parent table-responsive is-vertical">
            <!-- Token pages -->
            <nav class="pagination is-right">
              <a class="pagination-previous"
                v-on:click="loadPage(currentPage - 1)"
                v-bind:class="currentPage < 2 || loading ? 'is-disabled' : ''"
              >Previous</a>
              <a class="pagination-next"
                v-on:click="loadPage(currentPage + 1)"
                v-bind:class="currentPage > lastPage - 1 || loading ? 'is-disabled' : ''"
              >Next page</a>

              <ul class="pagination-list">
                <li>
                  <a class="pagination-link"
                    v-on:click="loadPage(1)"
                    v-bind:class="currentPage === 1 ? 'is-current' : ''"
                  >1</a>
                </li>
                <li v-if="currentPage > 3 && lastPage > 5">
                  <span class="pagination-ellipsis">&hellip;</span>
                </li>

                <li v-for="page in nearbyPages">
                  <a class="pagination-link"
                    v-on:click="loadPage(page)"
                    v-bind:class="page === currentPage ? 'is-current' : ''"
                  >{{ page }}</a>
                </li>

                <li v-if="lastPage - currentPage > 2 && lastPage > 5">
                  <span class="pagination-ellipsis">&hellip;</span>
                </li>
                <li>
                  <a class="pagination-link"
                    v-on:click="loadPage(lastPage)"
                    v-bind:class="currentPage === lastPage ? 'is-current' : ''"
                  >{{ lastPage }}</a>
                </li>
              </ul>
            </nav>

            <!-- Tokens table -->
            <table class="table is-striped is-narrow">
              <thead>
                <tr>
                  <th></th>
                  <th v-for="key in tableColumns">
                    {{ key | capitalize }}
                  </th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(entry, index) in tableData">
                  <td width="34">
                  <span class="icon">
                    <a @click="openModalBasic(index)">
                      <i class="fa fa-info"></i>
                    </a>
                  </span>
                  </td>
                  <td v-for="key in tableColumns">
                    {{ entry[key] }}
                  </td>
                  <td width="34">
                  <span class="icon">
                    <a @click="openDeleteModal(index)">
                      <i class="fa fa-trash-o"></i>
                    </a>
                  </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Userpass tab -->
          <div v-if="tabName === 'userpass'" class="table-responsive">
            <table class="table is-striped is-narrow">
              <thead>
                <tr>
                  <th></th>
                  <th v-for="key in tableColumns">
                    {{ key }}
                  </th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(entry, index) in tableData">
                  <td width="34">
                    <span class="icon">
                      <a @click="openModalBasic(index)">
                        <i class="fa fa-info"></i>
                      </a>
                    </span>
                  </td>
                  <td v-for="key in tableColumns">
                    {{ entry[key] }}
                  </td>
                  <td width="34">
                    <span class="icon">
                      <a @click="openDeleteModal(index)">
                        <i class="fa fa-trash-o"></i>
                      </a>
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- AppRole tab -->
          <div v-if="tabName === 'approle'" class="table-responsive">
            <table class="table is-striped is-narrow">
              <thead>
                <tr>
                  <th></th>
                  <th v-for="key in tableColumns">
                    {{ key }}
                  </th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(entry, index) in tableData">
                  <td width="34">
                    <span class="icon">
                      <a @click="openModalBasic(index)">
                        <i class="fa fa-info"></i>
                      </a>
                    </span>
                  </td>
                  <td v-for="key in tableColumns">
                    {{ entry[key] }}
                  </td>
                  <td width="34">
                    <span class="icon">
                      <a @click="openDeleteModal(index)">
                        <i class="fa fa-trash-o"></i>
                      </a>
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Certificates tab -->
          <!-- To be implemented -->

        </article>
      </div>
    </div>

    <modal :visible="showModal" :title="selectedItemTitle" :info="selectedItemInfo" @close="closeModalBasic"></modal>

    <confirmModal :visible="showDeleteModal" :title="confirmDeletionTitle" :info="selectedItemInfo" @close="closeDeleteModal" @confirmed="deleteItem(selectedIndex)"></confirmModal>

  </div>
</template>

<script>
import { Tabs, TabPane } from './vue-bulma-tabs'
import Modal from './modals/InfoModal'
import ConfirmModal from './modals/ConfirmModal'

var TabNames = ['token', 'userpass', 'approle']
var TabColumns = [
  [
    'accessor',
    'display_name',
    'num_uses',
    'orphan',
    'policies',
    'ttl'
  ],
  [
    'Name',
    'TTL',
    'Max_TTL',
    'Policies'
  ],
  [
    'Roleid',
    'Policies',
    'Token_TTL',
    'Token_max_TTL',
    'Secret_id_TTL',
    'Secret_id_num_uses'
  ]
]

export default {
  components: {
    Tabs,
    TabPane,
    Modal,
    ConfirmModal
  },

  data () {
    return {
      csrf: '',
      tabName: 'token',
      tableData: [],
      tableColumns: [
        'Token_Accessor',
        'Display_Name',
        'Num_Uses',
        'Orphan',
        'Path',
        'Policies',
        'TTL'
      ],
      showModal: false,
      showDeleteModal: false,
      selectedIndex: -1,
      currentPage: 1,
      lastPage: 6,
      loading: false
    }
  },

  mounted: function () {
    this.switchTab(0)
    this.$http.get('/api/tokencount').then((response) => {
      this.lastPage = Math.ceil(response.data.result / 300)
      this.csrf = response.headers['x-csrf-token']
    })
    .catch((error) => {
      this.$onError(error)
    })
  },

  computed: {
    selectedItemTitle: function () {
      if (this.selectedIndex !== -1) {
        return String(this.tableData[this.selectedIndex][this.tableColumns[0]])
      }
      return ''
    },
    selectedItemInfo: function () {
      if (this.selectedIndex !== -1) {
        return 'This modal panel is under construction'
      }
      return ''
    },
    confirmDeletionTitle: function () {
      return 'Are you sure you want to delete this?'
    },

    // calculates which pages should be directly click-able
    nearbyPages: function () {
      // if less than 5 pages, return all loadable pages
      if (this.lastPage < 6) {
        return [...Array(this.lastPage).keys()].slice(2)
      }
      if (this.currentPage === 1 || this.currentPage === 2) {
        return [2, 3]
      } else if (this.currentPage + 3 > this.lastPage) {
        return [this.lastPage - 3, this.lastPage - 2, this.lastPage - 1]
      } else {
        return [this.currentPage - 1, this.currentPage, this.currentPage + 1]
      }
    }
  },

  filters: {
    capitalize: function (str) {
      return str.charAt(0).toUpperCase() + str.slice(1)
    }
  },

  methods: {
    switchTab: function (index) {
      // on swap, clear data and load new column names
      this.tableData = []
      this.tabName = TabNames[index]
      this.tableColumns = TabColumns[index]
      // populate new table data according to tab name
      this.$http.get('/api/users?type=' + this.tabName).then((response) => {
        this.tableData = response.data.result
        this.csrf = response.headers['x-csrf-token']
      })
      .catch((error) => {
        this.$onError(error)
      })
    },

    openModalBasic (index) {
      this.selectedIndex = index
      this.showModal = true
    },
    closeModalBasic () {
      this.selectedIndex = -1
      this.showModal = false
    },
    openDeleteModal (index) {
      this.selectedIndex = index
      this.showDeleteModal = true
    },
    closeDeleteModal () {
      this.selectedIndex = -1
      this.showDeleteModal = false
    },

    deleteItem (index) {
      this.$http.post('/api/users/revoke', {
        Type: this.tabName.toLowerCase(),
        ID: this.tableData[index][this.tableColumns[0]]
      }, {
        headers: {'X-CSRF-Token': this.csrf}
      })
      .then((response) => {
        this.closeDeleteModal()
        this.tableData.splice(index, 1)
        this.$notify({
          title: 'Success',
          message: 'Deletion successful',
          type: 'success'
        })
      })
      .catch((error) => {
        this.closeDeleteModal()
        this.$onError(error)
      })
    },

    loadPage: function (pageNumber) {
      this.currentPage = pageNumber
      this.loading = true
      this.$http.get('/api/users?type=token&offset=' + ((this.currentPage - 1) * 300).toString()).then((response) => {
        this.tableData = response.data.result
        this.csrf = response.headers['x-csrf-token']
        this.loading = false
      })
      .catch((error) => {
        this.$onError(error)
        this.loading = false
      })
    }
  }

}
</script>

<style scoped>
  .button {
    margin: 5px 0 0;
  }

  .control .button {
    margin: inherit;
  }

  .fa-trash-o {
    color: red;
  }

  .fa-info {
    color: lightskyblue;
  }
</style>
