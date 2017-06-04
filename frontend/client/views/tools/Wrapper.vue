<template>
  <div>
    <div class="tile is-ancestor">
      <div class="tile is-parent">
        <article class="tile is-child box">
          <label class="label">Token</label>
          <div class="field has-addons">
            <p class="control is-expanded">
              <input class="input" type="text"
                     placeholder="Paste token here to unwrap"
                     v-model=currToken
              >
            </p>
            <p class="control is-pulled-right">
            <a class="button is-primary"
            @click="unWrapToken()"
            :disabled="currToken === ''">
            <span>Unwrap</span>
            </a>
            </p>
          </div>
          <label class="label"> Data </label>
          <article class="tile is-child box">

            <!-- data table -->

            <div class="table-responsive">
              <table class="table is-striped is-narrow">

                <!-- header -->

                <thead>
                  <tr is-expandedß>
                    <th>Key</th>
                    <th>Value</th>
                    <th></th>
                  </tr>
                </thead>
                <!-- body -->
                <tbody>
                  <tr>
                    <!-- Editable key field -->
                    <td>
                      <p class="control">
                        <input class="input is-small" type="text" placeholder="">
                      </p>
                    </td>

                    <!-- Editable value field -->
                    <td>
                      <p class="control">
                        <input class="input is-small" type="text" placeholder="">
                      </p>
                    </td>

                    <td width="30">
                      <a @click="deleteItem(index)">
                        <span class="icon">
                          <i class="fa fa-times-circle"></i>
                        </span>
                      </a>
                    </td>

                  </tr>

                  <!-- new key value pair insertion row -->
                  <tr>
                    <td>
                      <p class="control">
                        <input
                        class="input is-small"
                        type="text"
                        placeholder="Add a key to wrap"
                        v-model="newKey"
                        v-bind:class="[
                        newKey === '' ? '' : 'is-success',
                        newKeyExists ? 'is-danger' : '']"
                        >
                      </p>
                    </td>
                    <td>

                      <p class="control">
                        <input
                        class="input is-small"
                        type="text"
                        placeholder="Add a value to wrap"
                        v-model="newValue"
                        v-bind:class="[newValue === '' ? '' : 'is-success']"
                        >
                      </p>
                    </td>
                  </tr>
                </tbody>
                <tfoot>

                </tfoot>
              </table>
            </div>
            <p class="control">
              <a class="button is-primary"
              @click="wrapData()"
              :disabled="tableData === ''">
              <span>Wrap</span>
            </a>
          </p>
          </article>
        </article>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  data () {
    return {
      csrf: '',
      tableData: [],
      currToken: ''
    }
  },

  mounted: function () {
    this.$notify({
      title: 'Under Construction',
      message: 'This page doesn\'t work.',
      type: 'warning'
    })
  },

  methods: {
    wrapData: function (argument) {
      // body...
    },

    unWrapToken: function (argument) {
      // body...
    },

    deleteItem: function (index) {
      this.tableData.splice(index, 1)
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
