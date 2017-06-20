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
              <li disabled><a>Certificates</a></li>
            </ul>
          </div>

          <!-- Tokens tab -->
          <div v-if="tabName === 'token'" class="tile is-parent table-responsive is-vertical">

            <!-- Token pages -->
            <nav class="pagination is-right">
              <!-- styling hack until level component plays nice with pagination -->
              <a class="pagination-previous"
                v-on:click="search.show = !search.show"
                :disabled="loading"
              >Search</a>
              <a class="pagination-previous"
                v-on:click="loadPage(currentPage - 1)"
                :disabled="loading || currentPage < 2 || !!search.searched"
              >Previous</a>
              <a class="pagination-next"
                v-on:click="loadPage(currentPage + 1)"
                :disabled="loading || currentPage > lastPage - 1 || !!search.searched"
              >Next page</a>

              <ul class="pagination-list">
                <li>
                  <a class="pagination-link"
                    v-on:click="loadPage(1)"
                    v-bind:class="currentPage === 1 ? 'is-current' : ''"
                    :disabled="!!search.searched"
                  >1</a>
                </li>
                <li v-if="currentPage > 3 && lastPage > 5">
                  <span class="pagination-ellipsis">&hellip;</span>
                </li>

                <li v-for="page in nearbyPages">
                  <a class="pagination-link"
                    v-on:click="loadPage(page)"
                    v-bind:class="page === currentPage ? 'is-current' : ''"
                    :disabled="!!search.searched"
                  >{{ page }}</a>
                </li>

                <li v-if="lastPage - currentPage > 2 && lastPage > 5 && lastPage !== 1">
                  <span class="pagination-ellipsis">&hellip;</span>
                </li>
                <li v-if="lastPage !== 1">
                  <a class="pagination-link"
                    v-on:click="loadPage(lastPage)"
                    v-bind:class="currentPage === lastPage ? 'is-current' : ''"
                    :disabled="!!search.searched"
                  >{{ lastPage }}</a>
                </li>
              </ul>
            </nav>

            <!-- spacing -->
            <label class="label"></label>

            <!-- Token search bar -->
            <article v-if="search.show" class="tile is-child box">
              <nav class="level">

                <!-- Search by name -->
                <div class="level-left">
                  <div class="level-item">
                    <p class="control">
                      <button class="button is-danger"
                      :class="loading ? 'is-loading' : ''"
                      :disabled="search.searched === 0"
                      @click="resetSearch()">
                        Reset
                      </button>
                    </p>
                  </div>
                </div>

                <div class="level-item">
                  <p v-if="search.searched !== 0" class="subtitle is-5">
                    Found <strong>{{ search.found }}</strong> matches out of <strong>{{ search.searched }}</strong> tokens
                  </p>
                  <p v-else class="subtitle is-5">
                    Displaying <strong>{{Math.min(tokenCount, 300)}}</strong> out of <strong>{{tokenCount}}</strong> tokens
                  </p>
                </div>

                <!-- Search by content -->
                <div class="level-right">
                  <div class="level-item">
                    <div class="field has-addons">
                      <p class="control has-icons-right">
                        <span class="select">
                          <select v-model="search.regex">
                          <option v-bind:value="false">Substring</option>
                          <option v-bind:value="true">Regex</option>
                          </select>
                        </span>
                      </p>
                      <p class="control">
                        <input class="input" type="text"
                        placeholder="Search all tokens"
                        v-model="search.str"
                        @keyup.enter="searchByString()">
                      </p>
                      <p class="control">
                        <button class="button is-info"
                        :class="loading ? 'is-loading' : ''"
                        :disabled="search.str === ''"
                        @click="searchByString()">
                          Search
                        </button>
                      </p>
                    </div>
                  </div>
                </div>

              </nav>
            </article>

            <!-- spacing -->
            <label class="label"></label>

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
                  <td v-if="entry" v-for="key in tableColumns">
                    {{ entry[key] }}
                  </td>
                  <td v-else>
                    ERROR: An invalid token-accessor has been found
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
    Modal,
    ConfirmModal
  },

  data () {
    return {
      csrf: '',
      tabName: 'token',
      tableData: [],
      tableColumns: [
        'accessor',
        'display_name',
        'num_uses',
        'orphan',
        'policies',
        'ttl'
      ],
      showModal: false,
      showDeleteModal: false,
      selectedIndex: -1,
      currentPage: 1,
      lastPage: 1,
      tokenCount: 0,
      loading: false,
      // when adding properties here,
      // be careful with reactivity (overwritten by switchTab())
      search: {
        show: false,
        str: '',
        found: 0,
        searched: 0,
        regex: false,
        regexp: null
      }
    }
  },

  mounted: function () {
    this.switchTab(0)
    this.$http.get('/api/users/csrf').then((response) => {
      this.csrf = response.headers['x-csrf-token']
    })
    .catch((error) => {
      this.$onError(error)
    })
    this.$http.get('/api/tokencount').then((response) => {
      this.tokenCount = response.data.result
      this.lastPage = Math.ceil(response.data.result / 300)
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
    },

    searchRegex: function () {
      return this.search.regex
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
      this.search = {
        show: false,
        str: '',
        found: 0,
        searched: 0,
        regex: false,
        regexp: null
      }
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
      if (pageNumber < 1 || pageNumber > this.lastPage || this.search.searched) {
        return
      }
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
    },

    itemContainsSearchString: function (item) {
      if (item) {
        for (var i = 0; i < this.tableColumns.length; i++) {
          if (item[this.tableColumns[i]].toString().includes(this.search.str)) {
            return true
          }
        }
      }
      return false
    },

    itemContainsRegexExpr: function (item) {
      if (item) {
        for (var i = 0; i < this.tableColumns.length; i++) {
          if (this.search.regexp.test(item[this.tableColumns[i].toString()])) {
            return true
          }
        }
      }
      return false
    },

    searchByString: function () {
      if (this.search.str === '') {
        return
      }
      this.tableData = []
      this.search.found = 0
      this.search.searched = 0
      this.loading = this.lastPage // each completed async call will decrement this until false
      this.search.regexp = new RegExp(this.search.str)

      // make an async call for each page
      for (var i = 0; i < this.lastPage; i++) {
        this.$http.get('/api/users?type=token&offset=' + (i * 300).toString()).then((response) => {
          var found = false
          if (this.search.regex) {
            found = response.data.result.filter(this.itemContainsRegexExpr)
          } else {
            found = response.data.result.filter(this.itemContainsSearchString)
          }
          this.search.found += found.length
          this.search.searched += response.data.result.length
          this.tableData = this.tableData.concat(found)
          this.loading = this.loading - 1 || false
        })
        .catch((error) => {
          this.$onError(error)
          this.loading = this.loading - 1 || false
        })
      }
    },

    resetSearch: function () {
      this.search.str = ''
      this.search.found = 0
      this.search.searched = 0
      this.loadPage(this.currentPage)
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
