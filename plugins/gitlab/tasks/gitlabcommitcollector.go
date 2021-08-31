package tasks

import (
	"fmt"
	"time"

	"github.com/merico-dev/lake/config"
	"github.com/merico-dev/lake/logger"
	lakeModels "github.com/merico-dev/lake/models"
	"github.com/merico-dev/lake/plugins/core"
	"github.com/merico-dev/lake/plugins/gitlab/models"
)

type ApiCommitResponse []struct {
	Title          string
	Message        string
	ProjectId      int
	ShortId        string `json:"short_id"`
	AuthorName     string `json:"author_name"`
	AuthorEmail    string `json:"author_email"`
	AuthoredDate   string `json:"authored_date"`
	CommitterName  string `json:"committer_name"`
	CommitterEmail string `json:"committer_email"`
	CommittedDate  string `json:"committed_date"`
	WebUrl         string `json:"web_url"`
	Stats          struct {
		Additions int
		Deletions int
		Total     int
	}
}

func createApiClient() *core.ApiClient {
	return core.NewApiClient(
		config.V.GetString("GITLAB_ENDPOINT"),
		map[string]string{
			"Authorization": fmt.Sprintf("Bearer %v", config.V.GetString("GITLAB_AUTH")),
		},
		10*time.Second,
		3,
	)
}

func CollectCommits(projectId int) error {
	gitlabApiClient := createApiClient()

	res, err := gitlabApiClient.Get(fmt.Sprintf("projects/%v/repository/commits?with_stats=true", projectId), nil, nil)
	if err != nil {
		return err
	}

	gitlabApiResponse := &ApiCommitResponse{}

	logger.Info("res", res)

	err = core.UnmarshalResponse(res, gitlabApiResponse)

	if err != nil {
		logger.Error("Error: ", err)
		return nil
	}

	for _, value := range *gitlabApiResponse {
		gitlabCommit := &models.GitlabCommit{
			Title:          value.Title,
			Message:        value.Message,
			ProjectId:      projectId,
			ShortId:        value.ShortId,
			AuthorName:     value.AuthorName,
			AuthorEmail:    value.AuthorEmail,
			AuthoredDate:   value.AuthoredDate,
			CommitterName:  value.CommitterName,
			CommitterEmail: value.CommitterEmail,
			CommittedDate:  value.CommittedDate,
			WebUrl:         value.WebUrl,
			Additions:      value.Stats.Additions,
			Deletions:      value.Stats.Deletions,
			Total:          value.Stats.Total,
		}
		err = lakeModels.Db.Create(&gitlabCommit).Error
	}

	if err != nil {
		logger.Error("Error: ", err)
	}
	return nil
}
