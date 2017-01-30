<template>
  <div>

    <div class="tile is-ancestor">
      <div class="tile is-parent">
        <article class="tile is-child box">
          <h4 class="title">Tokens</h4>
          <div class="table-responsive">
            <table class="table is-bordered is-striped is-narrow">
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
        </article>
      </div>
    </div>

  </div>
</template>

<script>
  export default {
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
          'Policies"',
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
