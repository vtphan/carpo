<template>
    <div>
      <b-card no-body>
        <b-tabs card>
          <b-tab active>
            <template #title>
              <div v-on:click="getSnapshotList()">  <a v-if="message.data">({{ message.data.length}})</a> </div>
            </template>
              <div v-if="isLoading">
                <div>LOADING...</div>
              </div>
              <b-card-text v-else>
                <div>
                <v-row class="five-cols">
                <!-- <div class="items" > -->
                    <b-card
                      class="item"
                      v-b-modal = "'myModal2'"
                      :style="{'border-color': setborderColor(items.created_at)}"
                      v-for="items in message.data" :key="items.id"
                      @click="sendInfo(items)">
                        <template #header>
                          <div class="box-header d-flex justify-content-between align-items-center">
                            {{ items.id }} : {{ items.problem_id }}
                            <b-icon v-if="items.on_watch" icon="flag-fill" scale="2"></b-icon>
                          </div>
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
              <b-modal id="myModal2" size="xl" modal-class="custom-modal-size" :hide-footer="true" @shown="onModalShown">
                <template #modal-title>
                  <div class="box-header d-flex justify-content-between align-items-center">
                    <div style="margin-right: 20px;"> Snapshot {{ timeDiff(selectedSub.created_at) }} ago </div>
                    <b-icon v-if="selectedSub.on_watch" icon="flag-fill" scale="2"></b-icon>
                  </div>
                </template>
                <b-row class="resizable-container">
                  <b-col :style="{flex: `0 0 ${leftColumnWidth}%`}" class="resizable-column">
                    <h5>Code Snapshot</h5>
                    <codemirror ref="cmEditor" v-model="selectedSub.code" :options="cmOptions" />
                    <b-row class="mt-3">
                      <b-col cols="6">
                        <div style="text-align: left">
                          <div v-if="selectedSub.on_watch" class="row">
                              <b-button class="btn-secondary" @click="unwatchSub(selectedSub);">Unwatch</b-button>
                          </div>
                          <div v-else class="row">
                            <b-button class="btn-secondary" @click="watchSubmission(selectedSub);">Watch</b-button>
                            <b-form-input style="width: 60%; height: auto;" v-model="reason" placeholder="Reason to set on Watch (Optional)"></b-form-input>
                          </div>
                        </div>
                      </b-col>
                      <b-col cols="6">
                        <div style="text-align: right">
                          <b-button-group>
                            <b-button class="btn-secondary" @click="sendFeedback(selectedSub, selectedSub.id)">Send Feedback</b-button>
                          </b-button-group>
                        </div>
                      </b-col>
                    </b-row>
                  </b-col>
                  <div class="resize-handle" @mousedown="startResize"></div>
                  <b-col :style="{flex: `0 0 ${rightColumnWidth}%`}" class="resizable-column">
                    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px;">
                      <h5>Model Feedbacks</h5>
                      <b-button
                        variant="primary"
                        size="sm"
                        @click="fetchSecondColumnData"
                        :disabled="isLoadingSecondColumn"
                      >
                        {{ isLoadingSecondColumn ? 'Loading...' : 'Get AI Feedback' }}
                      </b-button>
                    </div>
                    <div v-if="isLoadingSecondColumn" class="text-center">
                      <b-spinner variant="primary"></b-spinner>
                      <p>Fetching AI feedbacks...</p>
                    </div>
                    <div v-else-if="secondColumnError" class="alert alert-warning">
                      {{ secondColumnError }}
                    </div>
                    <div v-else class="second-column-content">
                      <div v-if="feedbackList && feedbackList[selectedSub.id] && feedbackList[selectedSub.id].length > 0">
                        <div v-for="(feedback, index) in feedbackList[selectedSub.id]" :key="index" class="feedback-box">
                          <div class="feedback-header" style="display: flex; align-items: center; justify-content: flex-start; gap: 10px;">
                            <b-button
                              size="sm"
                              variant="outline-primary"
                              @click="appendFeedbackToCode(feedback.feedback)"
                              v-b-tooltip.hover
                              title="Append to Code"
                            >
                            <font-awesome-icon icon="arrow-rotate-left" />
                            </b-button>
                            <strong>Feedback {{ index + 1 }} ({{feedback.model}})</strong>
                          </div>
                          <div class="feedback-content">
                            <div v-if="feedback.feedback" class="feedback-message">{{ feedback.feedback }}</div>
                          </div>
                        </div>
                      </div>
                      <div v-else class="no-feedback">
                        <p>No feedback available</p>
                      </div>
                    </div>
                  </b-col>
                </b-row>
              </b-modal>
            </b-card-text>
          </b-tab>
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

import * as Config from '../config'
import moment from 'moment'

export default {
  name: 'SnapshotList',
  components: {
    codemirror
  },
  data: () => ({
    message: '',
    selectedSub: '',
    watchSubs: '',
    reason: '',
    sorting: 'creation_time',
    watchedSub: '',
    isLoading: true,
    secondColumnData: '',
    feedbackList: {},
    isLoadingSecondColumn: false,
    secondColumnError: '',
    leftColumnWidth: 48,
    rightColumnWidth: 48,
    isResizing: false,
    cmOptions: {
      autoRefresh: true,
      tabSize: 4,
      styleActiveLine: true,
      lineNumbers: true,
      line: true,
      mode: 'application/x-httpd-python',
      lineWrapping: true,
      theme: 'duotone-light'
    }
  }),
  methods: {
    sendInfo (item) {
      this.selectedSub = item
      this.reason = ''
    },
    getImagePath () {
      return require('../assets/code-block-1.png')
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
    watchSubmission (sub) {
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      }
      let postBody = {
        'student_id': sub.student_id,
        'submission_id': sub.id,
        'problem_id': sub.problem_id,
        'reason': this.reason,
        'mode': 1
      }

      this.$http.post(Config.apiUrl + '/snapshots/watch', postBody, config)
        .then(() => {
          // alert('Snapshot  with id ' + sub.student_id + ' is on watch list.')
          this.toast('Snapshot of student with id ' + sub.student_id + ' is on watch list.')
          this.message.data = this.message.data.map(obj => {
            if (obj.id === sub.id) {
              return { ...obj, on_watch: 1 }
            }
            return obj
          })
          this.selectedSub.on_watch = 1
        })
        .catch(function (error) {
          console.log(error)
          // alert(error)
        })
    },
    unwatchSub (sub) {
      console.log(sub)
      this.$http.delete(Config.apiUrl + '/snapshots/watch', {
        headers: { Authorization: 'Bearer ' + this.$route.query.token },
        data: {id: sub.watch_id}
      })
        .then(() => {
          // alert('Snapshot  with id ' + sub.student_id + ' is removed from the watch list.')
          this.toast('Snapshot of student with id ' + sub.student_id + ' is removed from the watch list.')
          this.message.data = this.message.data.map(obj => {
            if (obj.id === sub.id) {
              return { ...obj, on_watch: 0 }
            }
            return obj
          })
          this.selectedSub.on_watch = 0
        })
        .catch(function (error) {
          console.log(error)
          // alert(error)
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
          this.toast('Feedback is sent to student.')
        })
    },
    getSnapshotList: function () {
      this.isLoading = true
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token },
        params: {'sort_by': this.sorting}
      }
      this.$http.get(Config.apiUrl + '/snapshots/teachers', config)
        .then((response) => {
          console.log('Snapshot: ', response)
          this.message = response.data
          this.isLoading = false
        })
        .catch((error) => {
          console.log('Error', error)
          this.toast('Unauthorized Access.')
        })
    },
    getWatchedSubsList: function () {
      this.isLoading = true
      const config = {
        headers: { Authorization: 'Bearer '.concat(this.$route.query.token) },
        params: {'sort_by': this.sorting}
      }
      this.$http.get(Config.apiUrl + '/snapshots/watch', config)
        .then((response) => {
          this.watchSubs = response.data
          // console.log('watched' + JSON.stringify(this.watchSubs))
          this.isLoading = false
        })
        .catch((error) => {
          console.log('Error', error)
          this.toast('Unauthorized Access.')
        })
    },
    setSorting (params) {
      this.sorting = params
      this.getSnapshotList()
      this.getWatchedSubsList()
    },
    fetchSecondColumnData () {
      // Check if feedback already exists for this submission
      if (this.feedbackList[this.selectedSub.id] && this.feedbackList[this.selectedSub.id].length > 0) {
        return
      }
      this.isLoadingSecondColumn = true
      this.secondColumnError = ''
      this.secondColumnData = ''
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      }
      this.$http.get(Config.apiUrl + '/submissions/' + this.selectedSub.id + '/agent-feedback', config)
        .then((response) => {
          this.secondColumnData = response.data
          if (response.data && response.data.data && Array.isArray(response.data.data)) {
            const agentFeedbacks = response.data.data
            // Process each agent's feedback
            agentFeedbacks.forEach(agentData => {
              if (agentData.feedbacks && Array.isArray(agentData.feedbacks)) {
                const submissionId = agentData.submission_id || this.selectedSub.id
                if (!this.feedbackList[submissionId]) {
                  this.feedbackList[submissionId] = []
                }
                agentData.feedbacks.forEach(feedbackItem => {
                  this.feedbackList[submissionId].push({
                    feedback: feedbackItem.feedback || feedbackItem,
                    model: agentData.model || 'Unknown Model',
                    submission_id: submissionId
                  })
                })
              }
            })
          } else {
            this.feedbackList = {}
            this.secondColumnError = 'No feedback data available'
          }
          this.isLoadingSecondColumn = false
        })
        .catch((error) => {
          console.log('Error fetching agent feedback:', error)
          this.secondColumnError = 'Failed to load agent feedback'
          this.isLoadingSecondColumn = false
        })
    },
    formatTimestamp (timestamp) {
      if (!timestamp) return ''
      return moment(timestamp).format('MMM DD, YYYY HH:mm')
    },
    onModalShown () {
      // Refresh CodeMirror when modal is shown to fix display issue
      this.$nextTick(() => {
        if (this.$refs.cmEditor && this.$refs.cmEditor.codemirror) {
          this.$refs.cmEditor.codemirror.refresh()
        }
      })
    },
    appendFeedbackToCode (feedback) {
      if (feedback && this.selectedSub) {
        // Add feedback as a comment to the end of the code
        const feedbackComment = `\n\n# ${feedback.split('\n').join('\n# ')}`
        this.selectedSub.code = this.selectedSub.code + feedbackComment
        // Refresh CodeMirror to show the updated content
        this.$nextTick(() => {
          if (this.$refs.cmEditor && this.$refs.cmEditor.codemirror) {
            this.$refs.cmEditor.codemirror.refresh()
          }
        })
        this.toast('Feedback appended to code')
      }
    },
    startResize (event) {
      this.isResizing = true
      const startX = event.clientX
      const startLeftWidth = this.leftColumnWidth

      const handleMouseMove = (e) => {
        if (!this.isResizing) return

        const container = document.querySelector('.resizable-container')
        const containerRect = container.getBoundingClientRect()
        const containerWidth = containerRect.width

        const deltaX = e.clientX - startX
        const deltaPercent = (deltaX / containerWidth) * 100

        let newLeftWidth = startLeftWidth + deltaPercent

        // Constrain between 20% and 80%
        newLeftWidth = Math.max(20, Math.min(80, newLeftWidth))

        this.leftColumnWidth = newLeftWidth
        this.rightColumnWidth = 98 - newLeftWidth
      }

      const handleMouseUp = () => {
        this.isResizing = false
        document.removeEventListener('mousemove', handleMouseMove)
        document.removeEventListener('mouseup', handleMouseUp)
      }

      document.addEventListener('mousemove', handleMouseMove)
      document.addEventListener('mouseup', handleMouseUp)
      event.preventDefault()
    }
  },
  created: function () {
    this.getSnapshotList()
    // this.getWatchedSubsList()
    // setInterval(() => this.getSnapshotList(), 10000)
  }
}
</script>
<style>
/* https://gist.github.com/gokulkrishh/242e68d1ee94ad05f488 */
@media (min-width: 700px) {
  /* CSS */
  .five-cols {
    grid-template-columns: repeat(10, 1fr);
    column-gap: 8px;
    gap: 10px;
  }
}

.five-cols {
  display: grid;
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

.box-header {
  margin: 5px;
  padding-left: 10px;
  padding-right: 10px;
}

.btn-group, .btn-group-vertical {
    position: relative;
    display: -ms-inline-flexbox;
    display: -webkit-inline-box;
    /* display: inline-flex; */
    vertical-align: middle;
}

.custom-modal-size .modal-dialog {
    max-width: 95vw;
    width: 95vw;
}

.second-column-content {
    max-height: 600px;
    overflow-y: auto;
    background-color: #f8f9fa;
    padding: 15px;
    border-radius: 5px;
    border: 1px solid #dee2e6;
}

.feedback-box {
    background-color: #ffffff;
    border: 1px solid #e0e0e0;
    border-radius: 8px;
    margin-bottom: 15px;
    padding: 15px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    transition: box-shadow 0.3s ease;
}

.feedback-box:hover {
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
}

.feedback-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;
    padding-bottom: 8px;
    border-bottom: 1px solid #f0f0f0;
}

.feedback-header strong {
    color: #343a40;
    font-size: 16px;
}

.feedback-timestamp {
    color: #6c757d;
    font-size: 12px;
    font-style: italic;
}

.feedback-content {
    color: #495057;
}

.feedback-message {
    margin-bottom: 10px;
    line-height: 1.5;
}

.feedback-score {
    background-color: #e7f3ff;
    padding: 5px 10px;
    border-radius: 15px;
    display: inline-block;
    font-size: 12px;
    font-weight: bold;
    color: #0066cc;
    margin-bottom: 10px;
}

.feedback-suggestions {
    margin-top: 10px;
}

.feedback-suggestions ul {
    margin-left: 0;
    padding-left: 20px;
}

.feedback-suggestions li {
    margin-bottom: 5px;
    color: #495057;
}

.no-feedback {
    text-align: center;
    color: #6c757d;
    font-style: italic;
    padding: 40px 20px;
}

.resizable-container {
    display: flex !important;
    align-items: stretch;
    min-height: 600px;
    width: 100%;
}

.resizable-column {
    min-width: 0;
    padding: 0 15px;
}

.resize-handle {
    width: 8px;
    background-color: #ddd;
    cursor: col-resize;
    position: relative;
    z-index: 10;
    flex-shrink: 0;
    transition: background-color 0.2s ease;
}

.resize-handle:hover {
    background-color: #007bff;
}

.resize-handle::before {
    content: '';
    position: absolute;
    top: 50%;
    left: 2px;
    width: 4px;
    height: 20px;
    background-color: #666;
    transform: translateY(-48%);
    border-radius: 2px;
}
</style>
