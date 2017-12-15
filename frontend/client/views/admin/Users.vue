<template>
  <div>

    <div class="tile is-ancestor">
      <div class="tile is-parent is-vertical">
        <article class="tile is-child box">

          <!-- Tab navigation -->
          <div class="tabs is-medium is-boxed is-fullwidth">
            <ul>
              <li v-bind:class="tabName === 'token' ? 'is-active' : ''"
                v-on:click="switchTab(0, false)"
                disabled>
                <a>Tokens</a>
              </li>
              <li v-bind:class="tabName === 'userpass' ? 'is-active' : ''"
                v-on:click="switchTab(1, true)"
                :disabled="loading">
                <a>Userpass</a>
              </li>
              <li v-bind:class="tabName === 'approle' ? 'is-active' : ''"
                v-on:click="switchTab(2, true)"
                :disabled="loading">
                <a>Approle</a>
              </li>
              <li v-bind:class="tabName === 'ldap' ? 'is-active' : ''"
                v-on:click="switchTab(3, true)"
                :disabled="loading">
                <a>LDAP</a>
              </li>
            </ul>
          </div>

          <!-- Tokens tab -->
          <div v-if="tabName === 'token'" class="tile is-parent is-vertical">

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
                    Displaying <strong>{{Math.min(tableData.length, 300)}}</strong> out of <strong>{{accessors.length}}</strong> tokens
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
            <table class="table is-fullwidth is-striped is-narrow">
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
                      <span class="icon has-text-info">
                        <i class="fa fa-info-circle"></i>
                      </span>
                    </a>
                  </span>
                  </td>
                  <td v-if="entry" v-for="key in tableColumns">
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

            <a v-if="tableData.length === 0" class="pagination-next"
              v-on:click="switchTab(0, true)"
              :disabled="loading"
            >Load the first page of tokens</a>

          </div>

          <!-- Userpass tab -->
          <div v-if="tabName === 'userpass'">
            <table class="table is-fullwidth is-striped is-narrow">
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
                        <span class="icon has-text-info">
                          <i class="fa fa-info-circle"></i>
                        </span>
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
          <div v-if="tabName === 'approle'">
            <table class="table is-fullwidth is-striped is-narrow">
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
                        <span class="icon has-text-info">
                          <i class="fa fa-info-circle"></i>
                        </span>
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

          <!-- LDAP tab -->
          <div v-if="tabName === 'ldap'">

            <nav class="level">
              <div class="level-item has-text-centered">
                <div>
                  <p class="title is-4">Groups</p>
                </div>
              </div>
              <div class="level-item has-text-centered">
                <div>
                  <p class="title is-4">Users</p>
                </div>
              </div>
            </nav>

            <div class="columns">
              <div class="column">

                <!-- LDAP Groups table-->
                <table class="table is-fullwidth is-striped is-narrow">
                  <thead>
                    <tr>
                      <!-- If entry doesn't have 'Groups' field, it's a LDAP group -->
                      <th v-for="key in tableColumns" v-if="key != 'Groups'">
                        {{ key }}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    <!-- Ignore any entries with 'Groups' field -->
                    <tr v-for="(entry, index) in tableData" v-if="!entry.hasOwnProperty('Groups')">
                      <td v-for="key in tableColumns" v-if="key != 'Groups'">
                        {{ entry[key] }}
                      </td>
                    </tr>
                  </tbody>
                </table>

              </div>
              <div class="column">

                <table class="table is-fullwidth is-striped is-narrow">
                  <thead>
                    <tr>
                      <th v-for="key in tableColumns">
                        {{ key }}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    <!-- If the entry has 'Groups' field, it's an LDAP user -->
                    <tr v-for="(entry, index) in tableData" v-if="entry.hasOwnProperty('Groups')">
                      <td v-for="key in tableColumns">
                        {{ entry[key] }}
                      </td>
                    </tr>
                  </tbody>
                </table>

              </div>
            </div>
          </div>

        </article>
      </div>
    </div>

    <modal
      :visible="showModal"
      :title="selectedItemTitle"
      :info="selectedItemInfo"
      :infoIsJSON="true"
      @close="closeModalBasic">
    </modal>

    <confirmModal
      :visible="showDeleteModal"
      :title="confirmDeletionTitle"
      :info="selectedItemInfo"
      @close="closeDeleteModal"
      @confirmed="deleteItem(selectedIndex)">
    </confirmModal>

  </div>
</template>

<script>
import Modal from './modals/InfoModal'
import ConfirmModal from './modals/ConfirmModal'

var TabNames = ['token', 'userpass', 'approle', 'ldap']

export default {
  components: {
    Modal,
    ConfirmModal
  },

  data () {
    return {
      tabName: 'token',
      tableData: [],
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
    this.switchTab(0, false)
  },

  computed: {
    session: function () {
      return this.$store.getters.session
    },

    tableColumns: function () {
      switch (this.tabName) {
        case 'token': {
          return [
            'accessor',
            'display_name',
            'num_uses',
            'orphan',
            'policies',
            'ttl'
          ]
        }
        case 'userpass': {
          return [
            'Name',
            'TTL',
            'Max_TTL',
            'Policies'
          ]
        }
        case 'approle': {
          return [
            'Roleid',
            'Policies',
            'Token_TTL',
            'Token_max_TTL',
            'Secret_id_TTL',
            'Secret_id_num_uses'
          ]
        }
        case 'ldap': {
          return [
            'Name',
            'Policies',
            'Groups'
          ]
        }
      }
    },

    selectedItemTitle: function () {
      if (this.selectedIndex !== -1) {
        return 'Details'
      }
      return ''
    },
    selectedItemInfo: function () {
      if (this.selectedIndex !== -1) {
        return JSON.stringify(this.tableData[this.selectedIndex], null, '\t')
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
    // if fetchDetails is set to false, accessor details will not be fetched
    // this lightens potentially unnecessary stress on the vault server
    switchTab: function (index, fetchDetails = true) {
      // switching during loading is disabled
      if (this.loading) {
        return
      }
      this.loading = true

      // on swap, clear data and load new column names
      this.tableData = []
      this.tabName = TabNames[index]
      this.search = {
        show: false,
        str: '',
        found: 0,
        searched: 0,
        regex: false,
        regexp: null
      }

      // listing tokens
      if (this.tabName === 'token') {
        this.$http.get('/v1/token/accessors', {
          headers: {'X-Vault-Token': this.session ? this.session.token : ''}
        }).then((response) => {
          this.accessors = response.data.result
          this.lastPage = Math.ceil(this.accessors.length / 300)
          if (fetchDetails) {
            this.loadPage(1) // loadPage will turn loading to false
          } else {
            this.loading = false
          }
        })
        .catch((error) => {
          this.loading = false
          this.$onError(error)
        })

      // listing userpass users
      } else if (this.tabName === 'userpass') {
        this.$http.get('/v1/userpass/users', {
          headers: {'X-Vault-Token': this.session ? this.session.token : ''}
        }).then((response) => {
          this.loading = false
          this.tableData = response.data.result
        })
        .catch((error) => {
          this.loading = false
          this.$onError(error)
        })

      // listing approle roles
      } else if (this.tabName === 'approle') {
        this.$http.get('/v1/approle/roles', {
          headers: {'X-Vault-Token': this.session ? this.session.token : ''}
        }).then((response) => {
          this.loading = false
          this.tableData = response.data.result
        })
        .catch((error) => {
          this.loading = false
          this.$onError(error)
        })

      // listing ldap groups/users
      } else if (this.tabName === 'ldap') {
        // both ldap groups and ldap users will be in one array
        this.tableData = []

        // fetch ldap groups
        this.$http.get('/v1/ldap/groups', {
          headers: {'X-Vault-Token': this.session ? this.session.token : ''}
        }).then((response) => {
          this.loading = false
          this.tableData = this.tableData.concat(response.data.result)
        })
        .catch((error) => {
          this.loading = false
          this.$onError(error)
        })

        // fetch ldap users
        this.$http.get('/v1/ldap/users', {
          headers: {'X-Vault-Token': this.session ? this.session.token : ''}
        }).then((response) => {
          this.loading = false
          this.tableData = this.tableData.concat(response.data.result)
        })
        .catch((error) => {
          this.loading = false
          this.$onError(error)
        })

      // this should not be reachable through the UI by normal means
      } else {
        this.loading = false
        this.$notify({
          title: 'Invalid',
          message: 'Unsupported tab name',
          type: 'warning'
        })
      }
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
      // deleting a token via accessor
      if (this.tabName === 'token') {
        this.$http.post('/v1/token/revoke-accessor?accessor=' + this.tableData[index][this.tableColumns[0]], {}, {
          headers: {'X-Vault-Token': this.session ? this.session.token : ''}
        }).then((response) => {
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

      // deleting a user via username
      } else if (this.tabName === 'userpass') {
        this.$http.post('/v1/userpass/delete?username=' + encodeURIComponent(this.tableData[index][this.tableColumns[0]]), {}, {
          headers: {'X-Vault-Token': this.session ? this.session.token : ''}
        }).then((response) => {
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

      // deleting an approle via role name
      } else if (this.tabName === 'approle') {
        this.$http.post('/v1/approle/delete?role=' + encodeURIComponent(this.tableData[index][this.tableColumns[0]]), {}, {
          headers: {'X-Vault-Token': this.session ? this.session.token : ''}
        }).then((response) => {
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

      // this should not be reachable through the UI by normal means
      } else {
        this.$notify({
          title: 'Invalid',
          message: 'Unsupported tab name',
          type: 'warning'
        })
      }
    },

    loadPage: function (pg) {
      if (pg < 1 || pg > this.lastPage || this.search.searched) {
        return
      }
      this.currentPage = pg
      this.loading = true
      this.tableData = []

      // construct accessor string delimited by comma, and send search request
      this.$http.post('/v1/token/lookup-accessor', {
        Accessors: this.accessors.slice((pg - 1) * 300, pg * 300).join(',')
      }, {
        headers: {'X-Vault-Token': this.session ? this.session.token : ''}
      })
      .then((response) => {
        this.tableData = response.data.result
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
        this.$http.post('/v1/token/lookup-accessor', {
          Accessors: this.accessors.slice(i * 300, (i + 1) * 300).join(',')
        }, {
          headers: {'X-Vault-Token': this.session ? this.session.token : ''}
        })
        .then((response) => {
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

  } // end methods

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
</style>
