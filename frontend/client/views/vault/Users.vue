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
                    <tr v-for="entry in tableData">
                      <td class="is-icon">
                        <a href="#">
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
                    <tr v-for="entry in tableData">
                      <td class="is-icon">
                        <a href="#">
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

  </div>
</template>

<script>
  import { Tabs, TabPane } from './vue-bulma-tabs'

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
      TabPane
    },
    data () {
      return {
        searchQuery: '',
        tableData: [],
        tableColumns: [
          'Token_Accessor',
          'Display_Name',
          'Num_Uses',
          'Orphan',
          'Path',
          'Policies',
          'TTL'
        ]
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
