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
                  <tr>
                    <th>Key</th>
                    <th>Value</th>
                    <th width="1"></th>
                  </tr>
                </thead>
                <!-- body -->
                <tbody>
                  <tr v-for="(entry, index) in tableData">
                    <!-- Editable key field -->
                    <td v-if="entry.isClicked">
                      <p class="control">
                        <input class="input is-small" type="text" placeholder="" v-model="entry.key"
                               @keyup.enter="doneEdit(index)">
                      </p>
                    </td>

                    <!-- View-only -->
                    <td v-else @click="clicked(index)">
                      {{ entry.key }}
                    </td>

                    <!-- Editable value field -->
                    <td v-if="entry.isClicked">
                      <p class="control">
                        <input class="input is-small" type="text" placeholder="" v-model="entry.value"
                               @keyup.enter="doneEdit(index)">
                      </p>
                    </td>

                    <!-- View-only -->
                    <td v-else @click="clicked(index)">
                      {{ entry.value }}
                    </td>


                    <td width="30">
                      <a v-if="entry.isClicked" @click="deleteItem(index)">
                        <span class="icon">
                          <i class="fa fa-times-circle"></i>
                        </span>
                      </a>
                    </td>

                  </tr>

                  <!-- new key value pair insertion row -->
                  <tr @keyup.enter="addKeyValue()">
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

                <!-- footer only shows beyond a certain amount of data -->
                <tfoot v-show="tableData.length > 10">
                  <tr>
                    <th>Key</th>
                    <th>Value</th>
                    <th></th>
                  </tr>
                </tfoot>
              </table>
            </div>
            <nav class="level">
              <div class="level-left"></div>
              <div class="level-right">
                <p class="control">
                  <a class="button is-primary"
                     @click="WrapToken()"
                     :disabled="tableData.length === 0">
                  <span> Wrap </span>
                  </a>
                </p>
              </div>
            </nav>
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
      currToken: '',
      newKey: '',
      newValue: '',
      currKey: '',
      currVal: ''
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
    },

    clicked: function (index) {
      this.tableData[index].isClicked = true
    },

    addKeyValue: function () {
      // only allow insertion if key and value are valid
      if (this.newKey === '' || this.newValue === '') {
        this.$notify({
          title: 'Invalid',
          message: 'Key and Value must be non-empty',
          type: 'warning'
        })
        return
      }

      // insert new key value pair to table data
      this.tableData.push({
        key: this.newKey,
        value: this.newValue,
        isClicked: false
      })
      // reset so that a new pair can be inserted
      this.newKey = ''
      this.newValue = ''
    },

    doneEdit: function (index) {
      // check key and value again
      if (this.tableData[index].key === '' || this.tableData[index].value === '') {
        this.$notify({
          title: 'Invalid',
          message: 'Edits can\'t cause non-empty Key or Value ',
          type: 'warning'
        })
        return
      }
      this.tableData[index].isClicked = false
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
