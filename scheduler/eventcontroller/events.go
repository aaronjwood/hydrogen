package eventcontroller

/*
Adapted from mesos-framework-sdk
*/
import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"log"
	"mesos-framework-sdk/include/mesos"
	sched "mesos-framework-sdk/include/scheduler"
	"mesos-framework-sdk/resources"
	"mesos-framework-sdk/resources/manager"
	"mesos-framework-sdk/scheduler"
	"mesos-framework-sdk/task_manager"
	"mesos-framework-sdk/utils"
	"strconv"
)

type SprintEventController struct {
	scheduler       *scheduler.DefaultScheduler
	taskmanager     *task_manager.DefaultTaskManager
	resourcemanager *manager.DefaultResourceManager
	events          chan *sched.Event
	numOfExecutors  int
}

func NewSprintEventController(scheduler *scheduler.DefaultScheduler, manager *task_manager.DefaultTaskManager, resourceManager *manager.DefaultResourceManager, eventChan chan *sched.Event, e int) *SprintEventController {
	return &SprintEventController{
		taskmanager:     manager,
		scheduler:       scheduler,
		events:          eventChan,
		resourcemanager: resourceManager,
		numOfExecutors:  e,
	}
}

// Getter functions
func (s *SprintEventController) Scheduler() *scheduler.DefaultScheduler {
	return s.scheduler
}

// Getter function
func (s *SprintEventController) TaskManager() *task_manager.DefaultTaskManager {
	return s.taskmanager
}

// Getter function
func (s *SprintEventController) ResourceManager() *manager.DefaultResourceManager {
	return s.resourcemanager
}

func (s *SprintEventController) Subscribe(*sched.Event_Subscribed) {
	fmt.Println("Subscribe event.")
}

func (s *SprintEventController) Run() {
	if s.scheduler.FrameworkInfo().GetId() == nil {
		err := s.scheduler.Subscribe(s.events)
		if err != nil {
			log.Printf("Error: %v", err.Error())
		}

		select {
		case e := <-s.events:
			id := e.GetSubscribed().GetFrameworkId()
			s.scheduler.Info.Id = id
			log.Printf("Subscribed with an ID of %s", id.GetValue())
		}
		s.launchExecutors(s.numOfExecutors)
	}
	s.Listen()
}

// Create n default executors and launch them.
func (s *SprintEventController) launchExecutors(num int) {
	for i := 0; i < num; i++ {
		id, _ := utils.UuidToString(utils.Uuid())
		// Add tasks to task manager
		task := &mesos_v1.Task{
			Name:   proto.String("Sprint_" + id),
			TaskId: &mesos_v1.TaskID{Value: proto.String(id)},
			Resources: []*mesos_v1.Resource{
				resources.CreateCpu(0.1, "*"),
				resources.CreateMem(128.0, "*"),
			},
			State: mesos_v1.TaskState_TASK_STAGING.Enum(),
		}
		s.taskmanager.Add(task)
	}
}

// Main event loop that listens on channels forever until framework terminates.
func (s *SprintEventController) Listen() {
	for {
		select {
		case t := <-s.events:
			switch t.GetType() {
			case sched.Event_SUBSCRIBED:
				log.Println("Subscribe event.")
			case sched.Event_ERROR:
				go s.Error(t.GetError())
			case sched.Event_FAILURE:
				go s.Failure(t.GetFailure())
			case sched.Event_INVERSE_OFFERS:
				go s.InverseOffer(t.GetInverseOffers())
			case sched.Event_MESSAGE:
				go s.Message(t.GetMessage())
			case sched.Event_OFFERS:
				log.Println("Offers...")
				go s.Offers(t.GetOffers())
			case sched.Event_RESCIND:
				go s.Rescind(t.GetRescind())
			case sched.Event_RESCIND_INVERSE_OFFER:
				go s.RescindInverseOffer(t.GetRescindInverseOffer())
			case sched.Event_UPDATE:
				go s.Update(t.GetUpdate())
			case sched.Event_HEARTBEAT:
				fmt.Println("Heart beat.")
			case sched.Event_UNKNOWN:
				fmt.Println("Unknown event recieved.")
			}
		}
	}
}

func (s *SprintEventController) Offers(offerEvent *sched.Event_Offers) {
	fmt.Println("Offers event recieved.")
	//Reconcile any tasks.
	var reconcileTasks []*mesos_v1.Task
	s.scheduler.Reconcile(reconcileTasks)

	var offerIDs []*mesos_v1.OfferID

	for num, offer := range offerEvent.GetOffers() {
		fmt.Printf("Offer number: %v, Offer info: %v\n", num, offer)
		offerIDs = append(offerIDs, offer.GetId())
	}

	// Check task manager for any active tasks.
	if s.taskmanager.HasQueuedTasks() {
		// Update our resources in the manager
		s.resourcemanager.AddOffers(offerEvent.GetOffers())
		tasksToLaunch := []*mesos_v1.Task{}

		// See which tasks still need to be launched.
		for item := range s.taskmanager.Tasks().Iterate() {
			t := item.Value.(mesos_v1.Task)
			if t.GetState() == mesos_v1.TaskState_TASK_STAGING {
				tasksToLaunch = append(tasksToLaunch, &t)
			}
		}

		for _, item := range tasksToLaunch {
			// See if we have resources.
			if s.resourcemanager.HasResources() {
				taskList := []*mesos_v1.TaskInfo{} // Clear it out every time.
				operations := []*mesos_v1.Offer_Operation{}
				mesosTask := item

				offer, err := s.resourcemanager.Assign(mesosTask)
				if err != nil {
					// It didn't match any offers.
					log.Println(err.Error())
				}

				t := &mesos_v1.TaskInfo{
					Name:    mesosTask.Name,
					TaskId:  mesosTask.TaskId,
					AgentId: offer.AgentId,
					Command: &mesos_v1.CommandInfo{
						User:  proto.String("root"),
						Value: proto.String("/bin/sleep 10"),
					},
					Resources: []*mesos_v1.Resource{
						resources.CreateCpu(0.1, ""),
						resources.CreateMem(64.0, ""),
					},
				}

				taskList = append(taskList, t)

				operations = append(operations, resources.LaunchOfferOperation(taskList))

				log.Printf("Launching task %v\n", taskList)
				s.scheduler.Accept(offerIDs, operations, nil)
			}
		}
	} else {
		var ids []*mesos_v1.OfferID
		for _, v := range offerEvent.GetOffers() {
			ids = append(ids, v.GetId())
		}
		// decline offers.
		fmt.Println("Declining offers.")
		s.scheduler.Decline(ids, nil)
		s.scheduler.Suppress()
	}
}

func (s *SprintEventController) Rescind(rescindEvent *sched.Event_Rescind) {
	fmt.Printf("Rescind event recieved.: %v\n", *rescindEvent)
	rescindEvent.GetOfferId().GetValue()
}

func (s *SprintEventController) Update(updateEvent *sched.Event_Update) {
	fmt.Printf("Update recieved for: %v\n", *updateEvent.GetStatus())

	id := s.taskmanager.Get(updateEvent.GetStatus().GetTaskId())
	s.taskmanager.SetTaskState(id, updateEvent.GetStatus().State)

	status := updateEvent.GetStatus()
	s.scheduler.Acknowledge(status.GetAgentId(), status.GetTaskId(), status.GetUuid())
}

func (s *SprintEventController) Message(msg *sched.Event_Message) {
	fmt.Printf("Message event recieved: %v\n", *msg)
}

func (s *SprintEventController) Failure(fail *sched.Event_Failure) {
	log.Println("Executor " + fail.GetExecutorId().GetValue() + " failed with status " + strconv.Itoa(int(fail.GetStatus())))
}

func (s *SprintEventController) Error(err *sched.Event_Error) {
	fmt.Printf("Error event recieved: %v\n", err)
}

func (s *SprintEventController) InverseOffer(ioffers *sched.Event_InverseOffers) {
	fmt.Printf("Inverse Offer event recieved: %v\n", ioffers)
}

func (s *SprintEventController) RescindInverseOffer(rioffers *sched.Event_RescindInverseOffer) {
	fmt.Printf("Rescind Inverse Offer event recieved: %v\n", rioffers)
}
