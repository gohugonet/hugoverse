package application

import (
	"encoding/json"
	contentEntity "github.com/gohugonet/hugoverse/internal/domain/content/entity"
	"github.com/gohugonet/hugoverse/internal/domain/content/valueobject"
	hostEntity "github.com/gohugonet/hugoverse/internal/domain/host/entity"
	"github.com/gohugonet/hugoverse/internal/domain/host/factory"
	"github.com/gohugonet/hugoverse/pkg/timestamp"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func DeployToNetlify(target string, deployment *valueobject.Deployment, domain *valueobject.Domain, token string) error {
	host, err := factory.NewHost(logger)
	if err != nil {
		return err
	}

	if deployment.IsNewDeployment() {
		siteID, err := host.Netlify.DeployNewNetlifySite(token, target, deployment.SiteName, domain.FullDomain())
		if err != nil {
			return err
		}
		deployment.SiteID = siteID

		return nil
	}

	_, err = host.Netlify.DeployExistingNetlifySite(token, target, deployment.SiteID)
	if err != nil {
		return err
	}

	return nil
}

func PreviewSiteRecycle(cs *contentEntity.Content, token string) {
	host, err := factory.NewHost(logger)
	if err != nil {
		logger.Errorf("Failed to create host when recycle preview sites: %v", err)
		return
	}

	// 创建一个定时器，每隔1小时触发一次
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop() // 确保在程序退出时停止定时器

	// 创建一个通道来接收系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM) // 捕获中断信号

	logger.Println("预览站点回收任务已启动，将每小时执行一次...")

	for {
		select {
		case t := <-ticker.C:
			logger.Println("任务执行于:", t)
			recyclePreviewSites(cs, host, token) // 执行回收逻辑
		case sig := <-sigChan:
			logger.Printf("接收到信号: %v，程序即将退出...\n", sig)
			cleanup() // 执行清理操作
			return
		}
	}
}

// recyclePreviewSites 执行预览站点的回收逻辑
func recyclePreviewSites(cs *contentEntity.Content, host *hostEntity.Host, token string) {
	ns := "Preview"
	all := cs.Repo.AllContent(ns)
	p, ok := cs.AllContentTypes()[ns]
	if !ok {
		logger.Printf("Type %s not supported", ns)
		return
	}

	for _, v := range all {
		post := p()
		err := json.Unmarshal(v, post)
		if err != nil {
			logger.Println("Error unmarshalling when recycling ", ns, err)
		}

		if preview, ok := post.(*valueobject.Preview); ok {
			t, err := timestamp.ConvertInt64ToTime(preview.Time())
			if err != nil {
				logger.Println("Error converting time when recycling ", ns, err)
			}
			if timestamp.IsOneHourPassed(t) {
				err := host.Netlify.DeleteNetlifySite(token, preview.SiteID)
				if err != nil {
					logger.Println("Error deleting from Netlify when recycling ", ns, err)
					continue
				}

				idStr := strconv.Itoa(preview.ItemID())
				if err := cs.DeleteContent(ns, idStr, ""); err != nil {
					logger.Println("Error deleting content when recycling ", ns, err)
				}
			}
		}
	}
}

// cleanup 执行程序退出前的清理操作
func cleanup() {
	logger.Println("执行清理操作...")
}
