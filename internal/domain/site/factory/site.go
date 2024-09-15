package factory

import (
	"github.com/bep/gitmap"
	"github.com/gohugonet/hugoverse/internal/domain/site"
	"github.com/gohugonet/hugoverse/internal/domain/site/entity"
	"github.com/gohugonet/hugoverse/internal/domain/site/valueobject"
	"github.com/gohugonet/hugoverse/pkg/loggers"
)

func New(services site.Services) *entity.Site {
	log := loggers.NewDefault()

	git, err := newGitInfo(services)
	if err != nil {
		log.Errorf("failed to create git repo: %s", err)
		git = nil
	}

	s := &entity.Site{
		ConfigSvc:    services,
		ContentSvc:   services,
		ResourcesSvc: services,
		LanguageSvc:  services,

		GitSvc: git,

		Template: nil,

		Publisher: &entity.Publisher{Fs: services.Publish()},

		Title:    "Hugoverse",                                              // TODO: Make configurable
		Author:   valueobject.NewAuthor("Hugoverse", "support@gohugo.net"), // TODO: Make configurable
		Compiler: valueobject.NewVersion("0.0.0"),                          // TODO: Make configurable

		URL: &entity.URL{
			Base:      services.BaseUrl(),
			Canonical: true,
		},
		Language: &entity.Language{
			LangSvc: services,
		},
		Navigation: &entity.Navigation{},

		Log: loggers.NewDefault(),
	}

	s.PrepareLazyLoads()

	return s
}

func newGitInfo(conf site.FsService) (*valueobject.GitMap, error) {
	workingDir := conf.WorkingDir()

	gitRepo, err := gitmap.Map(gitmap.Options{
		Repository:        workingDir,
		Revision:          "",
		GetGitCommandFunc: nil,
	})
	if err != nil {
		return nil, err
	}

	return &valueobject.GitMap{ContentDir: gitRepo.TopLevelAbsPath, Repo: gitRepo}, nil
}
