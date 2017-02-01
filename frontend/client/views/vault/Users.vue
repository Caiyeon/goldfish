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
                        {{ key }}
                      </th>
                      <th></th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="(entry, index) in tableData">
                      <td class="is-icon">
                        <a @click="openModalBasic(index)">
                          <i class="fa fa-info"></i>
                        </a>
                      </td>
                      <td v-for="key in tableColumns">
                        {{ entry[key] }}
                      </td>
                      <td class="is-icon">
                        <a href="#">
                          <i class="fa fa-trash-o"></i>
                        </a>
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
                      <td class="is-icon">
                        <a a @click="openModalBasic(index)">
                          <i class="fa fa-info"></i>
                        </a>
                      </td>
                      <td v-for="key in tableColumns">
                        {{ entry[key] }}
                      </td>
                      <td class="is-icon">
                        <a href="#">
                          <i class="fa fa-trash-o"></i>
                        </a>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </tab-pane>
            <tab-pane label="AppRole" disabled>Video Tab</tab-pane>
            <tab-pane label="Certificates" disabled>Document Tab</tab-pane>
          </tabs>

        </article>
      </div>
    </div>

    <modal :visible="showModal" :title="selectedItemTitle" :info="selectedItemInfo" @close="closeModalBasic"></modal>

  </div>
</template>

<script>
  import { Tabs, TabPane } from './vue-bulma-tabs'
  import Modal from './modals/InfoModal'

  var TabNames = ['token', 'userpass']
  var TabColumns = [
    [
      'Token_Accessor',
      'Display_Name',
      'Num_Uses',
      'Orphan',
      'Path',
      'Policies',
      'TTL'
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
      Modal
    },

    data () {
      return {
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
        selectedIndex: -1
      }
    },

    computed: {
      selectedItemTitle: function () {
        if (this.selectedIndex !== -1) {
          return this.tableData[this.selectedIndex][this.tableColumns[0]]
        }
        return ''
      },
      selectedItemInfo: function () {
        if (this.selectedIndex !== -1) {
          return this.tableData[this.selectedIndex][this.tableColumns[1]]
        }
        return ''
      }
    },

    methods: {
      switchTab: function (index) {
        // on swap, clear data and load new column names
        this.tableData = []
        this.tableColumns = TabColumns[index]
        // populate new table data according to tab name
        this.$http.post('/api/users?type=' + TabNames[index]).then(function (response) {
          this.tableData = response['data']['result']
        }, function (err) {
          console.log(err)
        })
      },

      openModalBasic (index) {
        this.selectedIndex = index
        this.showModal = true
      },
      closeModalBasic () {
        this.selectedIndex = -1
        this.showModal = false
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
