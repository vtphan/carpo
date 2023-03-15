<template>
    <div>
      <b-card no-body>
        <b-tabs justified card>
          <b-tab active>
            <template #title>
              <div>
                <b-row>
                  <b-col cols="6" ><div style="float:left;">Snapshot</div></b-col>
                  <b-col cols="6" >
                    <div style="float:right;">
                      <b-dropdown no-caret>
                        <template #button-content>
                          <b-icon icon="gear-fill" aria-hidden="true"></b-icon> Filter By
                        </template>
                        <b-dropdown-item href="#" @click="setSorting('creation_time')">LastActive At</b-dropdown-item>
                        <b-dropdown-item href="#" @click="setSorting('name')">Name</b-dropdown-item>
                      </b-dropdown>
                    </div>
                  </b-col>
                </b-row>
              </div>
            </template>
            <b-card-text>
              <div >
                <div class="items" >
                    <b-card
                      class="item"
                      v-b-modal = "'myModal2'"
                      v-bind:img-src="getImagePath()"
                      img-alt="Card image"
                      style="max-width: 20rem;"
                      img-top v-for="items in message.data" :key="items.id"
                      @click="sendInfo(items)">
                        <b-card-text >
                            From: {{ items.student_name }}
                        </b-card-text>
                        <template #footer>
                            <small class="text-muted">Last Active: {{ timeDiff(items.created_at) }} ago </small>
                        </template>
                      </b-card>
                  </div>
              </div>

              <b-modal id="myModal2" title="Snapshot View" ok-only ok-variant="secondary" ok-title="Send Feedback">
                <codemirror v-model="selectedSub.code" :options="cmOptions" />
                <div style="text-align: right">
                  <b-button-group>
                    <b-button class="btn-secondary" @click="sendFeedback(selectedSub)">Send Feedback</b-button>
                  </b-button-group>
                </div>
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
    sorting: 'creation_time',
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
      console.log('SendInfo:', item)
      this.selectedSub = item
    },
    getImagePath () {
      return require('../assets/code-block-1.png')
    },
    timeDiff (dbTimestamp) {
      return moment.duration(moment().diff(moment(dbTimestamp))).humanize()
      // https://stackoverflow.com/questions/18623783/get-the-time-difference-between-two-datetimes
    },
    sendFeedback (submission) {
      let postBody = {
        'student_id': submission.student_id,
        'submission_id': submission.id,
        'problem_id': submission.problem_id,
        'teacher_id': this.$route.query.id,
        'code': submission.code
      }

      this.$http.post(Config.apiUrl + '/teachers/feedbacks', postBody)
        .then(data => {
          alert('Feedback sent to student.')
        })
    },
    getSnapshotList: function () {
      this.$http.get(Config.apiUrl + '/teachers/snapshots', {
        params: {
          'name': this.$route.query.name,
          'id': this.$route.query.id,
          'sort_by': this.sorting
        }
      })
        .then((response) => {
          console.log('Snapshot: ', response)
          this.message = response.data
        })
        .catch(function (error) {
          console.log(error)
        })
    },
    setSorting (params) {
      this.sorting = params
      this.getSnapshotList()
    }
  },
  created: function () {
    this.getSnapshotList()
    setInterval(() => this.getSnapshotList(), 10000)
  }
}
</script>
<style>

/* https://stackoverflow.com/questions/59445065/stack-v-cards-within-n-columns */
.items {
  padding: 5px;
  text-align: left;
  background-color: rgb(206, 209, 212);
}

.item {
  background-color: lightgrey;
  display: inline-block;
  width: 100%;
  margin: 10px;
}

.tab-content .active {
    padding: 0px;
}
/* Make it responsive */
@media only screen and (max-width: 1000px) {
  .items {
    column-count: 4;
  }
}

@media only screen and (max-width: 600px) {
  .items {
    column-count: 3;
  }
}

@media only screen and (max-width: 400px) {
  .items {
    column-count: 2;
  }
}

@media only screen and (max-width: 100px) {
  .items {
    column-count: 1;
  }
}

</style>
