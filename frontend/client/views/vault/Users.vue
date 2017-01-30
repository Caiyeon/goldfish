<template>
  <div>

    <div class="tile is-ancestor">
      <div class="tile is-parent is-vertical">
        <article class="tile is-child box">

          <tabs type="boxed" :is-fullwidth="true" alignment="centered" size="medium">

            <tab-pane label="Tokens">
              <div class="table-responsive">
                <table class="table is-striped is-narrow">
                  <thead>
                    <tr>
                      <th v-for="key in gridColumns">
                        {{ key }}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="entry in gridData">
                      <td v-for="key in gridColumns">
                        {{ entry[key] }}
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </tab-pane>

            <tab-pane label="Userpass">Music Tab</tab-pane>
            <tab-pane label="AppRole">Video Tab</tab-pane>
            <tab-pane label="Certificates">Document Tab</tab-pane>
          </tabs>

        </article>
      </div>
    </div>

  </div>
</template>

<script>
  import { Tabs, TabPane } from 'vue-bulma-tabs'

  export default {
    components: {
      Tabs,
      TabPane
    },
    data () {
      return {
        searchQuery: '',
        gridData: [],
        gridColumns: [
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
    created: function () {
      this.$http.post('/api/users').then(function (response) {
        this.gridData = response['data']['tokens']
      }, function (err) {
        console.log(err)
      })
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
</style>
