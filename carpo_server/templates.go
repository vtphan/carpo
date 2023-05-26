package main

var STUDENT_SUBMISSION_STATUS_TEMPLATE = `
<!DOCTYPE html>
<html>
	<head>
	<title>Submission Status</title>
	<meta http-equiv="refresh" content="120" >
	<script src="https://kit.fontawesome.com/923539b4ee.js" crossorigin="anonymous"></script>
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/bulma/0.9.3/css/bulma.min.css" integrity="sha512-IgmDkwzs96t4SrChW29No3NXBIBv8baW490zk5aXvhCD8vuZM3yUSkbyTBcXohkySecyzIrUwiF/qV0cuPcL3Q==" crossorigin="anonymous" referrerpolicy="no-referrer" />
	<script src='https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.12.0-2/js/all.min.js'></script>
	<style>
		#box {
			width: 70%;
			border: 1px solid #CCC;
			box-shadow: 0 1px 5px #CCC;
			margin: 25px auto;
		}
		#box header {
			background: #f1f1f1;
			box-shadow: 0 1px 2px #888;
			padding: 10px;
		}
		#box h1, h2 {
			padding: 0;
			margin: 0;
			font-size: 18px;
			text-align: center;
		}
		th {
			text-align: left;
		}
		.modal-content {
			margin-top: 100px;
			width: 450px;
		  }
	</style>
	</head>
	<body>
	<div id="box" class="container">
		<header>
			<h1>Submission Status<h1>
			<h2 style="font-size: 16px">Name: {{ .Name }}</h2>
		</header>
		<table class="table is-striped is-fullwidth is-hoverable is-narrow">
			<thead>
				<tr>
					<th>ProblemID</th>
					<th>SubmissionID</th>
					<th>Submitted</th>
					<th>Grade</th>
					<th>Graded</th>
					<th>Feedback</th>
				</tr>
			</thead>

			<tbody>
			{{ range $i, $e := .Stats }}
			<tr>
				<td>{{ .ProblemID }}</td>
				<td> <a href='#' id='btn-{{$i}}' class="modal-button" onclick='openpop({{$i}}, {{ .Code }})'>{{ .SubmissionID }}</a></td>
				<td>{{ .Submitted }}</td>
				<td>{{ if eq .Score 0 }} Ungraded {{else if eq .Score 1}} Correct {{else if eq .Score 2}} Incorrect {{end}}</td>
				<td>{{ if .Score }} {{ .GradeAt }} {{ end }} </td>
				<td>{{ if .HasFeedback }} {{ .FeedbackAt }} {{ end }} </td>
			</tr>
			{{ end }}
			</tbody>
		</table>
	</div>
	<div class="modal">
		<div class="modal-background"></div>
		<div class="modal-content" style="width: 55%;">
		<div class='box'>
			<h1 class='title'>Your Submitted Code</h1>
			<pre id="code-sec"></pre>
		</div>
		</div>
		<button class="modal-close is-large" 
				aria-label="close">
		Model
		</button>
	</div>
	<script>
    // Bulma does not have JavaScript included,
    // hence custom JavaScript has to be
    // written to open or close the modal
    const modal = 
          document.querySelector('.modal');
    const btn = 
          document.querySelector('#btn')
    const close = 
          document.querySelector('.modal-close')
  
    // btn.addEventListener('click',
    //                      function () {
    //   modal.style.display = 'block'
    // })

	function openpop(i, code) {
		modal.style.display = 'block'
		document.getElementById("code-sec").innerHTML = code;
	}
  
    close.addEventListener('click',
                           function () {
      modal.style.display = 'none'
    })
  
    window.addEventListener('click',
                            function (event) {
      if (event.target.className === 
          'modal-background') {
        modal.style.display = 'none'
      }
    })
  </script>

	</body>
</html>
`

var PROBLEM_GRADE_STATUS_TEMPLATE = `
<!DOCTYPE html>
<html>
	<head>
	<title>Problem Status</title>
	<meta http-equiv="refresh" content="120" >
	<script src="https://kit.fontawesome.com/923539b4ee.js" crossorigin="anonymous"></script>
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/bulma/0.9.3/css/bulma.min.css" integrity="sha512-IgmDkwzs96t4SrChW29No3NXBIBv8baW490zk5aXvhCD8vuZM3yUSkbyTBcXohkySecyzIrUwiF/qV0cuPcL3Q==" crossorigin="anonymous" referrerpolicy="no-referrer" />
	<style>
		#modal {
			width: 70%;
			border: 1px solid #CCC;
			box-shadow: 0 1px 5px #CCC;
			margin: 25px auto;
		}
		#modal header {
			background: #f1f1f1;
			box-shadow: 0 1px 2px #888;
			padding: 10px;
		}
		#modal h1 {
			padding: 0;
			margin: 0;
			font-size: 18px;
			text-align: center;
		}
	</style>
	</head>
	<body>
	<div id="modal" class="container">
		<header><h1>Problem Status</h1></header>
		<table class="table is-striped is-fullwidth is-hoverable is-narrow">
			<thead>
				<tr>
					<th>ProblemID</th>
					<th>Ungraded </th>
					<th>Correct </th>
					<th>Incorrect </th>
					<th>Published Date </th>
					<th>Unpublished Date </th>
				</tr>
			</thead>

			<tbody>
			{{ range .Stats }}
			<tr>
				<td><a href="/problem_detail?problem_id={{.ProblemID}}"> {{ .ProblemID }} </a> </td>
				<td>{{ .Ungraded}}</td>
				<td>{{ .Correct }} </td>
				<td>{{ .Incorrect }} </td>
				<td>{{ .PublishedDate.Format "Jan 02, 2006 3:04:05 PM" }} </td>
				<td>{{ if eq .ProblemStatus 0 }} {{ .UnpublishedDated.Format "Jan 02, 2006 3:04:05 PM" }} {{ else if eq .ProblemStatus 1 }} <em>  {{ .ExpiresAt  }} </em> {{ end }} </td>
			</tr>
			{{ end }}
			</tbody>
		</table>
	</div>
	</body>
</html>
`

var PROBLEM_DETAIL_TEMPLATE = `
<!DOCTYPE html>
<html>
	<head>
		<title>Problem Detail</title>
		<meta http-equiv="refresh" content="120" >
		<script src="https://kit.fontawesome.com/923539b4ee.js" crossorigin="anonymous"></script>
		<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/bulma/0.9.3/css/bulma.min.css" integrity="sha512-IgmDkwzs96t4SrChW29No3NXBIBv8baW490zk5aXvhCD8vuZM3yUSkbyTBcXohkySecyzIrUwiF/qV0cuPcL3Q==" crossorigin="anonymous" referrerpolicy="no-referrer" />
		<style>
			.row {
				display: flex;
			}
			.code-col {
				float: left;
				width: 70%;
				padding: 0 10px;
			}
			.detail-col {
				float: left;
				width: 30%;
				padding: 0 10px;
			}
			#modal {
				width: 70%;
				border: 1px solid #CCC;
				box-shadow: 0 1px 5px #CCC;
				margin: 25px auto;
			}
			#modal header {
				background: #f1f1f1;
				box-shadow: 0 1px 2px #888;
				padding: 10px;
			}
			#modal h1 {
				padding: 0;
				margin: 0;
				font-size: 18px;
				text-align: center;
			}
		</style>
	</head>
	<body>
		<div id="modal" class="container">
			<header><h1><a href="/problems/status">Problem Status</a></h1></header>
			<div class="row">
				<div class="code-col">
					<h4> Question </h4>
					<pre><code>{{ .Question }}</code></pre>
				</div>

				<div class="detail-col">
					<h4> Status </h4>
					<table class="table is-fullwidth is-hoverable is-narrow">
						<thead>
							<tr>
								<th></th>
								<th></th>
								
							</tr>
						</thead>

						<tbody>
							<tr> 
								<td> ID: </td> 
								<td> {{ .Stats.ProblemID }} </td>
							</tr>
							<tr> 
								<td> Ungraded Submissions: </td> 
								<td>{{ .Stats.Ungraded}} </td>
							</tr>
							<tr> 
								<td> Correct Submissions: </td> 
								<td>{{ .Stats.Correct }}  </td>
							</tr>
							<tr> 
								<td> Incorrect Submissions: </td> 
								<td>{{ .Stats.Incorrect }} </td>
							</tr>
							<tr> 
								<td> Published At: </td> 
								<td>{{ .Stats.PublishedDate.Format "Jan 02, 2006 3:04:05 PM" }} </td>
							</tr>
							<tr> 
								<td> Unpublished At: </td> 
								<td>{{ if eq .Stats.ProblemStatus 0 }} {{ .Stats.UnpublishedDated.Format "Jan 02, 2006 3:04:05 PM" }} {{ else if eq .Stats.ProblemStatus 1 }} <em>  {{ .Stats.ExpiresAt  }} </em> {{ end }} </td>
							</tr>
						</tbody>
					</table>
				</div>
			</div>
		</div>
	</body>
</html>
`
