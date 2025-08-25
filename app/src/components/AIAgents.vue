<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h2>Agent Settings</h2>
      <b-button variant="primary" @click="showCreateModal">Create New Agent</b-button>
    </div>

    <b-table striped hover :items="agents" :fields="fields" responsive="sm">
      <template #cell(description)="data">
        <div class="text-truncate" style="max-width: 200px;" :title="data.item.description">
          {{ data.item.description }}
        </div>
      </template>
      <template #cell(prompt)="data">
        <div class="text-truncate" style="max-width: 250px;" :title="data.item.prompt">
          {{ data.item.prompt }}
        </div>
      </template>
      <template #cell(is_active)="data">
        <b-form-checkbox
          :checked="data.item.is_active === 1"
          @change="toggleActive(data.item, $event)"
        ></b-form-checkbox>
      </template>
      <template #cell(actions)="row">
        <div class="sub-action">
          <b-button size="sm" @click="viewAgent(row.item)" class="mr-2" v-b-tooltip.hover title="View Agent">
            <font-awesome-icon icon="eye" />
          </b-button>
          <b-button size="sm" @click="editAgent(row.item)" class="mr-2" v-b-tooltip.hover title="Edit Agent">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
              <path d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.39-.39-1.02-.39-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z"/>
            </svg>
          </b-button>
          <b-button size="sm" variant="danger" @click="showDeleteConfirm(row.item)" v-b-tooltip.hover title="Delete Agent">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
              <path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/>
            </svg>
          </b-button>
        </div>
      </template>
    </b-table>

    <!-- Create/Edit Modal -->
    <b-modal :id="agentModal.id" :title="agentModal.title" size="xl" @ok="saveAgent" @hidden="resetForm" ok-title="Save">
      <b-form>
        <b-form-group label="Name" label-for="agent-name">
          <b-form-input
            id="agent-name"
            v-model="form.name"
            type="text"
            placeholder="Enter agent name"
            required
          ></b-form-input>
        </b-form-group>

        <b-form-group label="Description" label-for="agent-description">
          <b-form-input
            id="agent-description"
            v-model="form.description"
            type="text"
            placeholder="Enter agent description"
            required
          ></b-form-input>
        </b-form-group>

        <b-form-group label="Prompt" label-for="agent-prompt">
          <b-form-textarea
            id="agent-prompt"
            v-model="form.prompt"
            placeholder="Enter agent prompt"
            rows="6"
            required
          ></b-form-textarea>
        </b-form-group>

        <b-form-group label="Model" label-for="agent-model">
          <b-form-select
            id="agent-model"
            v-model="form.model"
            :options="modelOptions"
            required
          ></b-form-select>
        </b-form-group>

        <b-form-group>
          <b-form-checkbox
            id="agent-active"
            :checked="form.is_active === 1"
            @change="form.is_active = $event ? 1 : 0"
          >
            Active
          </b-form-checkbox>
        </b-form-group>
      </b-form>
    </b-modal>

    <!-- View Modal -->
    <b-modal :id="viewModal.id" :title="viewModal.title" size="xl" ok-only>
      <div v-if="selectedAgent">
        <h5>Name:</h5>
        <p>{{ selectedAgent.name }}</p>
        <h5>Description:</h5>
        <p>{{ selectedAgent.description }}</p>
        <h5>Model:</h5>
        <p>{{ selectedAgent.model }}</p>
        <h5>Active:</h5>
        <p>{{ selectedAgent.is_active === 1 ? 'Yes' : 'No' }}</p>
        <h5>Prompt:</h5>
        <pre class="bg-light p-3 rounded">{{ selectedAgent.prompt }}</pre>
      </div>
    </b-modal>
  </div>
</template>

<script>
import * as Config from '../config'

export default {
  name: 'AIAgents',
  data: () => ({
    agents: [],
    selectedAgent: null,
    form: {
      id: null,
      name: '',
      description: '',
      prompt: '',
      model: '',
      is_active: 1
    },
    agentModal: {
      id: 'agent-modal',
      title: 'Create Agent'
    },
    viewModal: {
      id: 'view-modal',
      title: 'Agent Details'
    },
    fields: [
      { key: 'name', label: 'Name', sortable: true },
      { key: 'description', label: 'Description', sortable: true },
      { key: 'model', label: 'Model', sortable: true },
      { key: 'prompt', label: 'Prompt', sortable: false },
      { key: 'is_active', label: 'Active', sortable: true },
      { key: 'actions', label: 'Actions', sortable: false }
    ],
    modelOptions: [
      { value: '', text: 'Select a model' },
      { value: 'gpt-4', text: 'GPT-4' },
      { value: 'gpt-3.5-turbo', text: 'GPT-3.5 Turbo' },
      { value: 'claude-3-opus', text: 'Claude 3 Opus' },
      { value: 'claude-3-sonnet', text: 'Claude 3 Sonnet' },
      { value: 'claude-3-haiku', text: 'Claude 3 Haiku' }
    ]
  }),
  methods: {
    showCreateModal () {
      this.agentModal.title = 'Create Agent'
      this.resetForm()
      this.$root.$emit('bv::show::modal', this.agentModal.id)
    },
    editAgent (agent) {
      this.agentModal.title = 'Edit Agent'
      this.form = { ...agent }
      this.$root.$emit('bv::show::modal', this.agentModal.id)
    },
    viewAgent (agent) {
      this.selectedAgent = agent
      this.viewModal.title = `Agent: ${agent.name}`
      this.$root.$emit('bv::show::modal', this.viewModal.id)
    },
    saveAgent (event) {
      event.preventDefault()
      if (!this.form.name || !this.form.description || !this.form.prompt || !this.form.model) {
        this.toast('Please fill in all fields', 'danger')
        return
      }
      if (this.form.id) {
        this.updateAgent()
      } else {
        this.createAgent()
      }
    },
    createAgent () {
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      }
      const agentData = {
        name: this.form.name,
        description: this.form.description,
        prompt: this.form.prompt,
        model: this.form.model,
        is_active: this.form.is_active
      }
      this.$http.post(Config.apiUrl + '/agents', agentData, config)
        .then((response) => {
          this.agents.push(response.data)
          this.toast('Agent created successfully', 'success')
          this.$nextTick(() => {
            this.$root.$emit('bv::hide::modal', this.agentModal.id)
          })
        })
        .catch((error) => {
          console.log('Error creating agent:', error)
          this.toast('Failed to create agent', 'danger')
        })
    },
    updateAgent () {
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      }
      const agentData = {
        name: this.form.name,
        description: this.form.description,
        prompt: this.form.prompt,
        model: this.form.model,
        is_active: this.form.is_active
      }
      this.$http.put(Config.apiUrl + '/agents/' + this.form.id, agentData, config)
        .then((response) => {
          this.toast('Agent updated successfully', 'success')
          this.$nextTick(() => {
            this.$root.$emit('bv::hide::modal', this.agentModal.id)
            this.getAgents()
          })
        })
        .catch((error) => {
          console.log('Error updating agent:', error)
          this.toast('Failed to update agent', 'danger')
        })
    },
    showDeleteConfirm (agent) {
      this.$bvModal.msgBoxConfirm('Are you sure you want to delete this agent?', {
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
          this.deleteAgent(agent)
        }
      })
    },
    deleteAgent (agent) {
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      }
      this.$http.delete(Config.apiUrl + '/agents/' + agent.id, config)
        .then(() => {
          const index = this.agents.findIndex(a => a.id === agent.id)
          if (index !== -1) {
            this.agents.splice(index, 1)
          }
          this.toast('Agent deleted successfully', 'success')
        })
        .catch((error) => {
          console.log('Error deleting agent:', error)
          this.toast('Failed to delete agent', 'danger')
        })
    },
    resetForm () {
      this.form = {
        id: null,
        name: '',
        description: '',
        prompt: '',
        model: '',
        is_active: 1
      }
    },
    toggleActive (agent, isChecked) {
      const newActiveValue = isChecked ? 1 : 0
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      }
      agent.is_active = newActiveValue
      this.$http.put(Config.apiUrl + '/agents/' + agent.id, agent, config)
        .then(() => {
          agent.is_active = newActiveValue
          this.toast(`Agent ${agent.name} ${newActiveValue === 1 ? 'activated' : 'deactivated'}`, 'info')
        })
        .catch((error) => {
          console.log('Error updating agent status:', error)
          this.toast('Failed to update agent status', 'danger')
        })
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
    getAgents () {
      const config = {
        headers: { Authorization: 'Bearer '.concat(this.$route.query.token) }
      }
      this.$http.get(Config.apiUrl + '/agents', config)
        .then((response) => {
          this.agents = response.data.data || response.data || []
        })
        .catch((error) => {
          console.log('Error loading agents:', error)
          this.toast('Failed to load agents', 'danger')
        })
    }
  },
  created () {
    this.getAgents()
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
