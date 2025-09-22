<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h2>Notebooks (Assignment/Exams)</h2>
      <b-button variant="primary" @click="showCreateModal">Upload Assignment</b-button>
    </div>

    <b-table striped hover :items="notebooks" :fields="fields" responsive="sm">
      <template #cell(mode)="data">
        <b-badge :variant="data.item.mode === 1 ? 'success' : 'primary'">
          {{ data.item.mode === 1 ? 'Assignment' : 'Exam' }}
        </b-badge>
      </template>
      <template #cell(available_till)="data">
        {{ formatDate(data.item.available_till) }}
      </template>
      <template #cell(end_time)="data">
        {{ formatDate(data.item.end_time) }}
      </template>
      <template #cell(actions)="row">
        <div class="sub-action">
          <b-button size="sm" @click="viewNotebook(row.item)" class="mr-2" v-b-tooltip.hover title="View Notebook Details">
            <font-awesome-icon icon="eye" />
          </b-button>
          <b-button size="sm" variant="danger" @click="showDeleteConfirm(row.item)" v-b-tooltip.hover title="Delete Notebook">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
              <path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/>
            </svg>
          </b-button>
        </div>
      </template>
    </b-table>

    <!-- Upload Modal -->
    <b-modal :id="notebookModal.id" :title="notebookModal.title" size="lg" @ok="saveAssignment" @hidden="resetForm">
      <b-form>
        <b-form-group label="Title" label-for="assignment-title">
          <b-form-input
            id="assignment-title"
            v-model="form.title"
            type="text"
            placeholder="Enter assignment title"
            required
          ></b-form-input>
        </b-form-group>

        <b-form-group label="Mode" label-for="assignment-mode">
          <b-form-select
            id="assignment-mode"
            v-model="form.mode"
            :options="modeOptions"
            required
          ></b-form-select>
        </b-form-group>

        <b-form-group label="Available To" label-for="assignment-available-to">
          <b-form-input
            id="assignment-available-to"
            v-model="form.available_till"
            type="datetime-local"
            required
          ></b-form-input>
        </b-form-group>

        <b-form-group label="End Time" label-for="assignment-end-time">
          <b-form-input
            id="assignment-end-time"
            v-model="form.end_time"
            type="datetime-local"
            required
          ></b-form-input>
        </b-form-group>

        <b-form-group label="Notebook File (.ipynb)" label-for="assignment-file">
          <b-form-file
            id="assignment-file"
            v-model="form.file"
            accept=".ipynb"
            placeholder="Choose a .ipynb file..."
            drop-placeholder="Drop .ipynb file here..."
            required
          ></b-form-file>
        </b-form-group>
      </b-form>
    </b-modal>

    <!-- View/Edit Modal -->
    <b-modal :id="viewModal.id" :title="viewModal.title" size="xl" @ok="updateNotebook" @hidden="resetViewForm">
      <div v-if="selectedNotebook">
        <h5>Title:</h5>
        <p>{{ selectedNotebook.title }}</p>
        <h5>Mode:</h5>
        <p>{{ selectedNotebook.mode === 1 ? 'Assignment' : 'Exam' }}</p>
        <h5>Available To:</h5>
        <b-form-input
          v-model="viewForm.available_till"
          type="datetime-local"
          class="mb-3"
        ></b-form-input>
        <h5>End Time:</h5>
        <b-form-input
          v-model="viewForm.end_time"
          type="datetime-local"
          class="mb-3"
        ></b-form-input>
      </div>
      <template #modal-footer="{ ok, cancel }">
        <b-button variant="secondary" @click="cancel()">Cancel</b-button>
        <b-button variant="primary" @click="ok()">Update Notebook</b-button>
      </template>
    </b-modal>
  </div>
</template>

<script>
import * as Config from '../config'

export default {
  name: 'Notebooks',
  data: () => ({
    notebooks: [],
    selectedNotebook: null,
    form: {
      title: '',
      mode: 1,
      available_till: '',
      end_time: '',
      file: null
    },
    viewForm: {
      available_till: '',
      end_time: ''
    },
    notebookModal: {
      id: 'assignment-modal',
      title: 'Upload Notebooks'
    },
    viewModal: {
      id: 'view-modal',
      title: 'Notebook Details'
    },
    fields: [
      { key: 'title', label: 'Title', sortable: true },
      { key: 'mode', label: 'Mode', sortable: true },
      { key: 'available_till', label: 'Available To', sortable: true },
      { key: 'end_time', label: 'End Time', sortable: true },
      { key: 'actions', label: 'Actions', sortable: false }
    ],
    modeOptions: [
      { value: 1, text: 'Assignment' },
      { value: 2, text: 'Exam' }
    ]
  }),
  methods: {
    showCreateModal () {
      this.notebookModal.title = 'Upload Notebook'
      this.resetForm()
      this.$root.$emit('bv::show::modal', this.notebookModal.id)
    },
    viewNotebook (notebook) {
      this.selectedNotebook = notebook
      this.viewModal.title = notebook.title
      this.viewForm.available_till = this.formatDateForInput(notebook.available_till)
      this.viewForm.end_time = this.formatDateForInput(notebook.end_time)
      this.$root.$emit('bv::show::modal', this.viewModal.id)
    },
    saveAssignment (event) {
      event.preventDefault()
      if (!this.form.title || this.form.mode === null || !this.form.available_till ||
          !this.form.end_time || !this.form.file) {
        this.toast('Please fill in all fields and select a file', 'danger')
        return
      }
      this.uploadAssignment()
    },
    uploadAssignment () {
      const reader = new FileReader()
      reader.onload = (e) => {
        const formData = new FormData()
        formData.append('title', this.form.title)
        formData.append('mode', this.form.mode)
        formData.append('available_till', new Date(this.form.available_till).toISOString())
        formData.append('end_time', new Date(this.form.end_time).toISOString())
        formData.append('user_id', 2)
        formData.append('filecontent', this.form.file)
        const config = {
          headers: {
            Authorization: 'Bearer ' + this.$route.query.token
          }
        }
        this.$http.post(Config.apiUrl + '/notebooks', formData, config)
          .then((response) => {
            this.notebooks.push(response.data)
            this.toast('Notebook uploaded successfully', 'success')
            this.$nextTick(() => {
              this.$root.$emit('bv::hide::modal', this.notebookModal.id)
            })
          })
          .catch((error) => {
            console.log('Error uploading assignment:', error)
            this.toast('Failed to upload assignment', 'danger')
          })
      }
      reader.readAsText(this.form.file)
    },
    showDeleteConfirm (assignment) {
      this.$bvModal.msgBoxConfirm('Are you sure you want to delete this notebook?', {
        title: 'Confirm Delete',
        size: 'sm',
        buttonSize: 'sm',
        okVariant: 'danger',
        okTitle: 'Delete',
        cancelTitle: 'Cancel',
        footerClass: 'p-2',
        hideHeaderClose: false,
        centered: true
      }).then(value => {
        if (value) {
          this.deleteAssignment(assignment)
        }
      })
    },
    deleteAssignment (assignment) {
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      }
      this.$http.delete(Config.apiUrl + '/notebooks/' + assignment.id, config)
        .then(() => {
          const index = this.notebooks.findIndex(a => a.id === assignment.id)
          if (index !== -1) {
            this.notebooks.splice(index, 1)
          }
          this.toast('Notebook deleted successfully', 'success')
        })
        .catch((error) => {
          console.log('Error deleting Notebook:', error)
          this.toast(`Failed to delete Notebook: ${error.response.data.error}`, 'danger')
        })
    },
    resetForm () {
      this.form = {
        title: '',
        mode: 1,
        available_till: '',
        end_time: '',
        file: null
      }
    },
    formatDate (dateString) {
      if (!dateString) return ''
      return new Date(dateString).toLocaleString()
    },
    formatDateForInput (dateString) {
      if (!dateString) return ''
      const date = new Date(dateString)
      return date.toISOString().slice(0, 16)
    },
    updateNotebook (event) {
      event.preventDefault()
      if (!this.viewForm.available_till || !this.viewForm.end_time) {
        this.toast('Please fill in both timestamp fields', 'danger')
        return
      }
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      }
      const updateData = {
        available_till: new Date(this.viewForm.available_till).toISOString(),
        end_time: new Date(this.viewForm.end_time).toISOString()
      }
      this.$http.put(Config.apiUrl + '/notebooks/' + this.selectedNotebook.id, updateData, config)
        .then((response) => {
          // console.log(response)
          const index = this.notebooks.findIndex(n => n.id === this.selectedNotebook.id)
          if (index !== -1) {
            this.notebooks[index].available_till = updateData.available_till
            this.notebooks[index].end_time = updateData.end_time
          }
          this.toast(response.data.msg, 'success')
          this.$nextTick(() => {
            this.$root.$emit('bv::hide::modal', this.viewModal.id)
          })
        })
        .catch((error) => {
          console.log('Error updating notebook:', error)
          this.toast(`Failed to update notebook times: ${error.response.data.error}`, 'danger')
        })
    },
    resetViewForm () {
      this.viewForm = {
        available_till: '',
        end_time: ''
      }
    },
    toast (message, variant = 'secondary') {
      this.$bvToast.toast(message, {
        title: 'Notification',
        toaster: 'b-toaster-top-center',
        variant: variant,
        autoHideDelay: 2000,
        solid: true
      })
    },
    getNotebooks () {
      const config = {
        headers: { Authorization: 'Bearer '.concat(this.$route.query.token) }
      }
      this.$http.get(Config.apiUrl + '/notebooks', config)
        .then((response) => {
          this.notebooks = response.data.data || response.data.notebooks || []
          console.log('Notebook: ', this.notebooks)
        })
        .catch((error) => {
          console.log('Error loading notebooks:', error)
          this.toast('Failed to load notebooks', 'danger')
        })
    }
  },
  created () {
    this.getNotebooks()
  }
}
</script>

<style>
.sub-action {
  margin: 5px;
}

.text-truncate {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
