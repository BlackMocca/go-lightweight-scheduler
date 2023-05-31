# go-lightweight-cronjob
This project is cronjob scheduler with Workflow like mini airflow

## Feature
- สามารถสร้างระบบการทำงาน pipeline ขนาดเล็กได้
- สามารถสั่งให้ทำงานได้ทันที หรือ ตั้งเวลาล่วงหน้า
- สามารถกำหนดการทำงานเมื่อ สำเร็จ หรือ พบข้อผิดพลาด
- สามารถกำหนดปกป้อง API ด้วย basic auth หรือ apikey
- support workflow ที่ทำงานเฉพาะผ่าน API เท่านั้น

## Requirement
- Required database postgres or mongodb and you must create database before running app
- This program running with docker image

## Documentation
- [Wiki Here](https://github.com/BlackMocca/go-lightweight-scheduler/wiki)
- [Reference API DOCS Here](https://github.com/BlackMocca/go-lightweight-scheduler/wiki/API-DOCUMENT)

<hr>

### Example Write workflow in dags/newbie.go
```golang
func startDagExampleNewbie() {
	config := scheduler.NewDefaultSchedulerConfig()
	schedulerInstance := scheduler.NewScheduler("*/5 * * * *", "example_newbie", "ทดสอบ bash_executor", config)

	job := scheduler.NewJob(nil)
	job.AddTask(
		task.NewTask("runbash", executor.NewBashExecutor(`ls -al`, true)),
		task.NewTask("runscript", executor.NewBashExecutor(`./script/test_bash.sh`, true)),
	)

	schedulerInstance.RegisterJob(job)
	register(schedulerInstance)
}
```


### Add call startdagExampleNewbie() in call function 
```golang
func call() {
	if constants.ENV_ENABLED_DAG_EXAMPLE {
		startDagExampleGolang()
		startDagExampleTaskBash()
		startDagExampleTaskBranch()
		startDagExampleWorkWithoutCronjob()
	}
	//startdagExampleNewbie()
}
```

