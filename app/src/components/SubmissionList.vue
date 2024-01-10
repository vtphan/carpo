<template>
  <div>
    <!-- <div>
      <h2 style="float: left;"> Available Tags: </h2>
      <br clear="all"/>
      <b-list-group horizontal>
        <b-list-group-item v-for="items in available_tags.data" :key="items.id">
          <label><input type="checkbox" :value="items.id" v-model="filter_tag" v-on:click="filterList()"> <span class="checkbox-label"> {{items.name}} </span></label> <br>
        </b-list-group-item>
      </b-list-group>
      <span> Filter Tags: {{ filter_tag }}</span>
    </div> -->
    <div class="row" style="margin: 5px;">
      <div>
        <h4 style="margin: 5px;">Filter By Tag: </h4>
      </div>
      <div style="width: 50%;">
        <multiselect v-model="filter_tag" track-by="id" label="name" placeholder="Select" :options="available_tags.data" :multiple="true" :close-on-select="false" :clear-on-select="false" :searchable="true">
          <template slot="singleLabel" slot-scope="{ option }"><strong>{{ option.name }}</strong> </template>
        </multiselect>
      </div>
    </div>
    <div>

    </div>
    <b-card no-body>
      <b-tabs card>
        <b-tab active>
          <template #title>
            <div v-on:click="getSubmissionList()">Submission <a v-if="message.data">({{ message.data.length}})</a></div>
          </template>
          <!-- <div style="float:right; position: absolute; top: 6px; left: calc(100% - 165px);">
            <b-dropdown no-caret>
              <template #button-content>
                <b-icon icon="gear-fill" aria-hidden="true"></b-icon> Order By
              </template>
              <b-dropdown-item href="#" @click="setSorting('creation_time')">Creation Time</b-dropdown-item>
              <b-dropdown-item href="#" @click="setSorting('name')">Name</b-dropdown-item>
            </b-dropdown>
          </div> -->
          <div v-if="isLoading">
                <p>LOADING...</p>
          </div>
          <b-card-text v-else>
            <div >
              <v-row class="five-cols">
              <!-- <div class="items" > -->
                  <b-card
                    class="item"
                    :style="{'border-color': setborderColor(items.created_at)}"
                    v-b-modal = "'myModal'"
                    v-for="items in message.data" :key="items.id"
                    @click="sendInfo(items)">
                      <template #header>
                        {{ items.id }} : {{ items.problem_id }}
                      </template>
                      <b-card-text >
                          {{ items.student_name }}
                      </b-card-text>
                      <template #footer>
                          <small>
                            Active {{ timeDiff(items.created_at) }} ago
                          </small>
                      </template>
                  </b-card>
              <!-- </div> -->
              </v-row>
            </div>

            <b-modal id="myModal" size="xl" :hide-footer="true">
                <template #modal-title>
                  Submission
                  <b-badge v-if="selectedSub.score==1" variant="success">correct</b-badge>
                  <b-badge v-if="selectedSub.score==2" variant="danger">incorrect</b-badge>
                  <b-badge v-if="!selectedSub.score" variant="secondary">ungraded</b-badge>
                </template>
                <codemirror v-model="selectedSub.code" :options="cmOptions" :style="{ height: '600px' }" ref="focusThis" />
                <!-- <a> Message: {{ selectedSub.message }} </a> -->
                <b-row>
                  <b-col cols="6" >
                    <div style="text-align: left">
                      <!-- <div class="row">
                        <b-button class="btn-secondary" @click="flagSubmission(selectedSub);">Flag</b-button>
                        <b-form-input style="width: 60%; height: auto;" v-model="reason" placeholder="Reason to flag (Optional)"></b-form-input>
                      </div> -->
                      <div class="row" style="margin: 5px;">
                        <h4 style="margin: 5px;">Tag: </h4>
                        <multiselect style="width: 50%;" v-model="assign_tags" track-by="id" label="name" placeholder="Select one" :options="available_tags.data" @select="saveSubmissionTag" @remove="remove_tag" :multiple="true" :close-on-select="false" :clear-on-select="false" :searchable="false">
                          <template slot="singleLabel" slot-scope="{ option }"><strong>{{ option.name }}</strong> </template>
                        </multiselect>
                        <h5 class="new-tag-link" v-on:click="newTag()" > Create New Tag </h5>
                      </div>
                    </div>
                  </b-col>
                  <b-col cols="6" >
                    <div style="text-align: right">
                      <b-button-group>
                        <b-button class="btn-success" @click="sendGrade(selectedSub, selectedSub.id, 1); ">Correct</b-button>
                        <b-button class="btn-danger" @click="sendGrade(selectedSub, selectedSub.id, 2); ">Incorrect</b-button>
                        <b-button class="btn-secondary" @click="sendFeedback(selectedSub, selectedSub.id);">Send Feedback</b-button>
                        <b-button class="btn-secondary" @click="watchSubmission(selectedSub);">Watch</b-button>
                      </b-button-group>
                    </div>
                  </b-col>
                </b-row>
            </b-modal>
          </b-card-text>
        </b-tab>
        <!-- <b-tab >
          <template #title>
            <div v-on:click="getFlaggedSubsList()"> Flagged <a v-if="flagSubs.data">({{ flagSubs.data.length}})</a>
            </div>
          </template>
          <div v-if="isLoading">
                <p>LOADING...</p>
          </div>
          <b-card-text v-else>
            <div>
              <v-row class="five-cols">
                  <b-card
                    class="item"
                    v-b-modal = "'flagModal'"
                    :style="{'border-color': setborderColor(items.created_at)}"
                    v-for="items in flagSubs.data" :key="items.id"
                    @click="sendInfo(items)">
                      <template #header >
                        {{ items.submission_id }} : {{ items.problem_id }}
                      </template>
                      <b-card-text >
                          {{ items.student_name }}
                      </b-card-text>
                      <template #footer>
                          <small>
                            Active {{ timeDiff(items.created_at) }} ago
                            <br>
                            {{ items.reason }}
                          </small>
                      </template>
                  </b-card>
              </v-row>
            </div>
            <b-modal id="flagModal" size="xl" :hide-footer="true">
                 <template #modal-title>
                    Submission
                    <b-badge v-if="selectedSub.score==1" variant="success">correct</b-badge>
                    <b-badge v-if="selectedSub.score==2" variant="danger">incorrect</b-badge>
                    <b-badge v-if="!selectedSub.score" variant="secondary">ungraded</b-badge>
                    <b-badge v-if="selectedSub.reason" variant="info">Tag: {{ selectedSub.reason }}</b-badge>
                  </template>
                <codemirror id='code-section' v-model="selectedSub.code" :options="cmOptions" ref="focusThis" />
                <a> Message: {{ selectedSub.message }} </a>
                <b-row>
                  <b-col cols="6" >
                    <div style="text-align: left">
                      <b-button class="btn-secondary" @click="Unflag(selectedSub); ">Unflag</b-button>
                    </div>
                  </b-col>
                  <b-col cols="6" >
                    <div style="text-align: right">
                      <b-button-group>
                        <b-button class="btn-success" @click="sendGrade(selectedSub, selectedSub.submission_id, 1); ">Correct</b-button>
                        <b-button class="btn-danger" @click="sendGrade(selectedSub,selectedSub.submission_id, 2); ">Incorrect</b-button>
                        <b-button class="btn-secondary" @click="sendFeedback(selectedSub, selectedSub.submission_id); ">Send Feedback</b-button>
                      </b-button-group>
                    </div>
                  </b-col>
                </b-row>
            </b-modal>
          </b-card-text>
        </b-tab> -->
      </b-tabs>
    </b-card>
  </div>
</template>

<script>
import { codemirror } from 'vue-codemirror'
import 'codemirror/lib/codemirror.css'
// language
import 'codemirror/mode/python/python.js'

// theme css
import 'codemirror/theme/duotone-light.css'

import Multiselect from 'vue-multiselect'

import * as Config from '../config'
import moment from 'moment'

export default {
  name: 'SubmissionList',
  components: {
    codemirror,
    Multiselect
  },
  data: () => ({
    token: '',
    message: '',
    datas: '',
    flagSubs: '',
    reason: '',
    sorting: 'creation_time',
    selectedSub: '',
    isLoading: true,
    cmOptions: {
      autoRefresh: true,
      tabSize: 4,
      styleActiveLine: true,
      lineNumbers: true,
      line: true,
      mode: 'application/x-httpd-python',
      lineWrapping: true,
      theme: 'duotone-light'
    },
    available_tags: {
      data: []
    },
    filter_tag: [],
    assign_tags: [],
    f_tag: []
  }),
  methods: {
    sendInfo (item) {
      this.selectedSub = item
      this.reason = ''
      this.assign_tags = item.tag
    },
    close (sub) {
      this.$bvModal.hide()
      this.message.data = this.message.data.filter(item => item.id !== sub.id)
    },
    getImagePath () {
      return require('../assets/code-block-1.png')
    },
    focusMyElement () {
      this.$refs.focusThis.focus()
    },
    timeDiff (dbTimestamp) {
      return moment.duration(moment().diff(moment(dbTimestamp))).humanize()
      // https://stackoverflow.com/questions/18623783/get-the-time-difference-between-two-datetimes
    },
    getColor (value) {
      // value from 0 to 1
      var hue = ((1 - value) * 120).toString(10)
      return ['hsl(', hue, ',100%,50%)'].join('')
    },
    setborderColor (dbTimestamp) {
      // https://stackoverflow.com/questions/7128675/from-green-to-red-color-depend-on-percentage
      const limit = 5
      var ago = moment.duration(moment().diff(moment(dbTimestamp))).minutes()
      if (ago > 5) {
        ago = 5
      }
      var value = ago / limit
      return this.getColor(value)
    },
    toast (msg) {
      this.$bvToast.toast(`${msg}`, {
        title: `Notification`,
        toaster: 'b-toaster-top-center',
        variant: 'secondary',
        autoHideDelay: 2000,
        solid: true
      })
    },
    filterList () {
      console.log('Here: ', this.filter_tag)
    },
    flagSubmission (submission) {
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      }
      let postBody = {
        'student_id': submission.student_id,
        'submission_id': submission.id,
        'problem_id': submission.problem_id,
        'reason': this.reason,
        'mode': 2
      }

      this.$http.post(Config.apiUrl + '/submissions/flag', postBody, config)
        .then(() => {
          // alert('This submission is now flagged.')
          this.toast('Submission with id ' + submission.id + ' is flagged.')
          // this.message.data = this.message.data.filter(item => item.id !== submission.id)
        })
        .catch(function (error) {
          console.log(error)
        })
    },
    Unflag (submission) {
      this.$http.delete(Config.apiUrl + '/submissions/flag', {
        headers: { Authorization: 'Bearer ' + this.$route.query.token },
        data: {id: submission.id, submission_id: submission.submission_id}
      })
        .then(() => {
          // alert('This submission is now unflagged.')
          this.toast('Submission with id ' + submission.submission_id + ' is unflagged.')
          // this.flagSubs.data = this.flagSubs.data.filter(item => item.id !== submission.id)
        })
        .catch(function (error) {
          console.log(error)
          // alert(error)
        })
    },
    sendGrade (submission, id, score) {
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      }

      let postBody = {
        'student_id': submission.student_id,
        'submission_id': id,
        'problem_id': submission.problem_id,
        // 'teacher_id': Number(this.$route.query.id),
        'score': score,
        'code': submission.code
      }

      var status = score === 1 ? 'Correct.' : 'Incorrect.'
      submission.score = score

      this.$http.post(Config.apiUrl + '/submissions/grades', postBody, config)
        .then(data => {
          // alert('This submission is now graded as ' + status)
          this.toast('Submission with id ' + id + ' is graded as ' + status)
          // this.message.data = this.message.data.filter(item => item.id !== id)
          // this.flagSubs.data = this.flagSubs.data.filter(item => item.id !== submission.id)
        })
    },
    sendFeedback (submission, id) {
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      }

      let postBody = {
        'student_id': submission.student_id,
        'submission_id': id,
        'problem_id': submission.problem_id,
        'code': submission.code
      }

      this.$http.post(Config.apiUrl + '/submissions/grades', postBody, config)
        .then(data => {
          // alert('Feedback sent to student.')
          this.toast('Feedback for submission with id ' + id + ' is sent to student.')
          // this.message.data = this.message.data.filter(item => item.id !== id)
          // this.flagSubs.data = this.flagSubs.data.filter(item => item.id !== submission.id)
        })
    },
    watchSubmission (sub) {
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      }
      let postBody = {
        'student_id': sub.student_id,
        'submission_id': sub.id,
        'problem_id': sub.problem_id,
        'mode': 2
      }
      this.$http.post(Config.apiUrl + '/snapshots/watch', postBody, config)
        .then(() => {
          // alert('Snapshot  with id ' + sub.student_id + ' is on watch list.')
          this.toast('Student with id ' + sub.student_id + ' is on watch list.')
        })
        .catch(function (error) {
          console.log(error)
          // alert(error)
        })
    },
    getSubmissionList: function () {
      this.isLoading = true
      const config = {
        headers: { Authorization: 'Bearer '.concat(this.$route.query.token) },
        params: {'sort_by': this.sorting}
      }
      this.$http.get(Config.apiUrl + '/submissions/teachers', config)
        .then((response) => {
          // this.datas = structuredClone(response.data)
          this.datas = JSON.parse(JSON.stringify(response.data))
          this.message = response.data
          this.isLoading = false
        })
        .catch((error) => {
          console.log('Error', error)
          this.toast('Unauthorized Access.')
        })
    },
    getFlaggedSubsList: function () {
      this.isLoading = true
      const config = {
        headers: { Authorization: 'Bearer '.concat(this.$route.query.token) },
        params: {'sort_by': this.sorting}
      }
      console.log(config)
      this.$http.get(Config.apiUrl + '/submissions/flag', config)
        .then((response) => {
          this.flagSubs = response.data
          this.isLoading = false
        })
        .catch((error) => {
          console.log('Error', error)
          this.toast('Unauthorized Access.')
        })
    },
    getAvailableTags () {
      const config = {
        headers: { Authorization: 'Bearer '.concat(this.$route.query.token) }
      }
      this.$http.get(Config.apiUrl + '/tags?mode=2', config)
        .then((response) => {
          this.available_tags = response.data
          // console.log('Available Tags: ' + JSON.stringify(this.available_tags))
        })
        .catch((error) => {
          console.log('Error', error)
          this.toast('Unauthorized Access.')
        })
    },
    saveSubmissionTag ({ name, id }) {
      //  console.log("Select: ", name, id, this.selectedSub.id)
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      }
      let postBody = {
        'submission_id': this.selectedSub.id,
        'tag_id': id
      }
      this.$http.post(Config.apiUrl + '/tags/submissions/', postBody, config)
        .then((resp) => {
          console.log(resp)
          this.toast('Submission Tag is saved.')
        })
        .catch(function (error) {
          alert(error)
        })
      // update the submission list with new tag
      var subIndex
      // console.log(this.message.data)
      subIndex = this.message.data.findIndex(obj => obj.id === this.selectedSub.id)
      if (this.message.data[subIndex].tag === undefined) {
        this.message.data[subIndex].tag = []
      }
      this.message.data[subIndex].tag.push({'id': id, 'name': name})
    },
    remove_tag ({ name, id }) {
      // console.log('Removing: ', name, id, this.selectedSub.id)
      this.$http.delete(Config.apiUrl + '/tags/' + id + '/submissions/' + this.selectedSub.id, {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      })
        .then((resp) => {
          console.log(resp)
          this.toast('Tag for submission with id ' + this.selectedSub.id + ' is removed.')
        })
        .catch(function (error) {
          alert(error)
        })
      // update the submission list with removed tag
      var subIndex
      console.log(this.message.data)
      subIndex = this.message.data.findIndex(obj => obj.id === this.selectedSub.id)
      console.log('with Tag: ', this.message.data[subIndex].tag)
      this.message.data[subIndex].tag = this.message.data[subIndex].tag.filter(item => item.id !== id)
    },
    newTag () {
      var link = document.createElement('a')
      link.href = window.location.origin + '/#/tags?token=' + this.$route.query.token
      link.target = '_blank'
      link.click()
    },
    setSorting (params) {
      this.sorting = params
      // this.getSubmissionList()
      // this.getFlaggedSubsList()
    }
  },
  created: function () {
    this.getSubmissionList()
    this.getAvailableTags()
    // this.getFlaggedSubsList()
  },
  watch: {
    filter_tag: function (val, oldVal) {
      // console.log(val, oldVal)
      if (val.length === 0) {
        // console.log('No Tag selected.')
        this.message.data = this.datas.data
        return
      }
      var newArray = []
      var ids = []
      this.datas.data.forEach((sub) => {
        if (sub.tag !== null) {
          sub.tag.forEach((t) => {
            val.forEach((selected) => {
              if (selected.id === t.id & ids.indexOf(sub.id) === -1) {
                newArray.push(sub)
                ids.push(sub.id) // Avoid duplicate entries
              }
            })
          })
        }
      })
      this.message.data = newArray
    }
  }
}
</script>

<style>

.five-cols {
  display: grid;
  grid-template-columns: repeat(10, 1fr);
  column-gap: 8px;
  background-color: rgb(206, 209, 212);
  padding: 5px;
  /* text-align: left; */
}

.item .card-header {
  padding: 0.25rem 0.25rem;
  font-weight: 300;
  font-size: 12px;
}

.card-body {
  padding: 2px;
  height: 55px;
}

b-card-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-footer {
  padding: 0.25rem 0.25rem;
}

small, .small {
    font-size: 11px;
    font-weight: 400;
}

br {
  content: "";
  margin: -2em;
  display: block;
  font-size: 24%;
}

.item {
  background-color: rgb(235, 229, 229);
  display: inline-block;
  width: 100%;
  max-width: 14rem;
  margin: 5px;
  font-size: 12px;
  font-weight: 300;
  border-width: 3px;
}

.tab-content .active {
    padding: 0px;
}

button {
  margin: 5px
}

input:placeholder-shown {
   font-style: italic;
}

.CodeMirror {
  height: 600px;
}

.new-tag-link {
  margin: 10px;
  text-decoration: underline;
  cursor: pointer;
}

</style>

<style src="vue-multiselect/dist/vue-multiselect.min.css"></style>
