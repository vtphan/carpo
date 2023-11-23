<template>
    <div class='tabview'>
        <section v-if="isLoading">
            <p>LOADING...</p>
        </section>

        <div class='tag-form'>
          <b-form inline @submit="onSubmit">
            <label class="sr-only" for="inline-form-input-name">Name For Tag</label>
            <b-form-input v-model="tag_form.name"
              id="inline-form-input-name"
              class="mb-2 mr-sm-2 mb-sm-0 name-field"
              placeholder="Tag Name"
            ></b-form-input>

            <label class="mr-sm-2" for="inline-form-custom-select-usecase">Usecase</label>
            <b-form-select v-model="tag_form.mode"
              id="inline-form-custom-select-usecase"
              class="mb-2 mr-sm-2 mb-sm-0"
              :options="options"
              :value=0
            ></b-form-select>
            <b-button type="submit" variant="primary">Create Tag</b-button>
          </b-form>
        </div>
        <div class='sub-head'>
          <p> All Available Tags Status: </p>
        </div>
        <div class="table-view">
          <b-table striped hover :items="all_tags.data" :fields="fields" >
            <template #cell(actions)="row">
              <b-button size="sm"  @click="deleteTag(row.item)" class="mr-2">
                Delete
              </b-button>
            </template>
          </b-table>
        </div>
    </div>
</template>

<script>
import * as Config from '../config'
export default {
  name: 'Tag',
  data: () => ({
    tag_form: {
      name: '',
      mode: 0
    },
    options: [
      { value: 0, text: 'Select Usecase' },
      { value: 1, text: 'Problems' },
      { value: 2, text: 'Submissions/Students' }
    ],
    all_tags: '',
    fields: [
      {
        key: 'tag_id',
        label: 'Tag ID',
        sortable: true
      },
      {
        key: 'name',
        label: 'Tag Name',
        sortable: true
      },
      {
        key: 'mode',
        label: 'Mode/UseCase',
        formatter: value => {
          return value === 1 ? 'Problems' : 'Submissions'
        },
        sortable: true
      },
      {
        key: 'count',
        label: 'No. of Tags',
        sortable: true
      },
      'actions'
    ]
  }),
  methods: {
    toast (msg) {
      this.$bvToast.toast(`${msg}`, {
        title: `Notification`,
        toaster: 'b-toaster-top-center',
        variant: 'secondary',
        autoHideDelay: 2000,
        solid: true
      })
    },
    onSubmit (event) {
      event.preventDefault()
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      }
      if (this.tag_form.name === '' || this.tag_form.mode === 0) {
        alert('Invalid input.', this.tag_form)
        return
      }
      this.$http.post(Config.apiUrl + '/tags', JSON.stringify(this.tag_form), config)
        .then((res) => {
          this.toast('New Tag Created.')
          console.log('Tag: ', res.data)
        })
        .catch(function (error) {
          console.log(error)
          this.toast('Failed to create tag.')
        })
    },
    getAllTags () {
      this.isLoading = true
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      }
      this.$http.get(Config.apiUrl + '/tags/tagged', config)
        .then((response) => {
          this.all_tags = response.data
          this.isLoading = false
          console.log(response.data)
        })
        .catch((error) => {
          console.log('Error', error)
          this.toast('Unauthorized Access.')
        })
    },
    deleteTag (item) {
      if (item.count !== 0) {
        alert('Tag cannot be deleted. Remove usecase.')
        return
      }

      this.$http.delete(Config.apiUrl + '/tags/' + item.tag_id, {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      })
        .then(() => {
          this.toast('Tag with id ' + item.tag_id + ' is deleted.')
        })
        .catch(function (error) {
          console.log(error)
        })
    }
  },
  created: function () {
    this.getAllTags()
  }
}
</script>

<style>

.tabview{
  padding: 10px;
}

.tag-form{
  margin: auto;
  width: 50%;
  border: 3px solid #dbe3db;
  padding: 10px;
}

.table-view{
  margin: auto;
  width: 50%;
  padding: 10px;
}

.sub-head {
  margin-top: 15px;
  font-size: 30px;
}

.name-field {
  width: 55% !important;
}

</style>
