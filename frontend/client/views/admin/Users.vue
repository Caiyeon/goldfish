<template>
  <div>

    <div class="tile is-ancestor">
      <div class="tile is-parent is-vertical">
        <article class="tile is-child box">

          <tabs type="boxed" :is-fullwidth="true" alignment="centered" size="medium" v-on:switched="switchTab">

            <tab-pane label="Tokens">
              <div class="table-responsive">
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
            </tab-pane>

            <tab-pane label="Userpass">
              <div class="table-responsive">
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
            </tab-pane>

            <tab-pane label="AppRole" disabled>Disabled</tab-pane>
            <tab-pane label="Certificates" disabled>Disabled</tab-pane>

          </tabs>

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
  import Vue from 'vue'
  import Notification from 'vue-bulma-notification'

  const NotificationComponent = Vue.extend(Notification)

  const openNotification = (propsData = {
    title: '',
    message: '',
    type: '',
    direction: '',
    duration: 4500,
    container: '.notifications'
  }) => {
    return new NotificationComponent({
      el: document.createElement('div'),
      propsData
    })
  }

  var TabNames = ['token', 'userpass']
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
        selectedIndex: -1
      }
    },

    computed: {
      selectedItemTitle: function () {
        if (this.selectedIndex !== -1) {
          return String(this.tableData[this.selectedIndex][this.tableColumns[1]])
        }
        return ''
      },
      selectedItemInfo: function () {
        if (this.selectedIndex !== -1) {
          return this.tableData[this.selectedIndex][this.tableColumns[0]]
        }
        return ''
      },
      confirmDeletionTitle: function () {
        return 'Are you sure you want to delete this?'
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
        this.$http.get('/api/users?type=' + this.tabName)
          .then((response) => {
            this.tableData = response.data.result
            this.csrf = response.headers['x-csrf-token']
          })
          .catch((error) => {
            openNotification({
              title: 'Error',
              message: error.body.error,
              type: 'danger'
            })
            console.log(error.body.error)
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
        this.$http
          .post('/api/users/revoke', {
            Type: this.tabName.toLowerCase(),
            ID: this.tableData[index][this.tableColumns[0]]
          }, {
            headers: {'X-CSRF-Token': this.csrf}
          })

          .then((response) => {
            this.tableData.splice(index, 1)
            openNotification({
              title: 'Success',
              message: 'Deletion successful',
              type: 'success'
            })
          })

          .catch((error) => {
            openNotification({
              title: 'Error',
              message: error.body.error,
              type: 'danger'
            })
            console.log(error.body.error)
          })

        this.showDeleteModal = false
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
