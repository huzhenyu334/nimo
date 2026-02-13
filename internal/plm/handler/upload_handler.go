package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UploadHandler 文件上传处理器
type UploadHandler struct{}

// NewUploadHandler 创建文件上传处理器
func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

// UploadedFile 上传文件信息
type UploadedFile struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
}

// Upload 处理文件上传
// POST /upload
func (h *UploadHandler) Upload(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		BadRequest(c, "无法解析上传文件: "+err.Error())
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		// 也尝试获取单文件
		files = form.File["file"]
	}
	if len(files) == 0 {
		BadRequest(c, "没有上传文件")
		return
	}

	now := time.Now()
	dir := fmt.Sprintf("./uploads/%d/%02d", now.Year(), now.Month())

	// 创建目录
	if err := os.MkdirAll(dir, 0755); err != nil {
		InternalError(c, "创建上传目录失败: "+err.Error())
		return
	}

	var uploaded []UploadedFile

	for _, fileHeader := range files {
		fileID := uuid.New().String()[:32]
		ext := filepath.Ext(fileHeader.Filename)
		savedName := fmt.Sprintf("%s_%s%s", fileID, fileHeader.Filename, "")
		if ext != "" {
			savedName = fmt.Sprintf("%s_%s", fileID, fileHeader.Filename)
		}
		savePath := filepath.Join(dir, savedName)

		// 打开源文件
		src, err := fileHeader.Open()
		if err != nil {
			InternalError(c, "读取上传文件失败: "+err.Error())
			return
		}

		// 创建目标文件
		dst, err := os.Create(savePath)
		if err != nil {
			src.Close()
			InternalError(c, "保存文件失败: "+err.Error())
			return
		}

		_, err = io.Copy(dst, src)
		src.Close()
		dst.Close()
		if err != nil {
			InternalError(c, "写入文件失败: "+err.Error())
			return
		}

		url := fmt.Sprintf("/uploads/%d/%02d/%s", now.Year(), now.Month(), savedName)

		uploaded = append(uploaded, UploadedFile{
			ID:          fileID,
			URL:         url,
			Filename:    fileHeader.Filename,
			Size:        fileHeader.Size,
			ContentType: fileHeader.Header.Get("Content-Type"),
		})
	}

	Success(c, uploaded)
}

// Get3DModel 获取3D模型的STL格式（用于前端Three.js预览）
// GET /files/:fileId/3d
func (h *UploadHandler) Get3DModel(c *gin.Context) {
	fileID := c.Param("fileId")
	if fileID == "" {
		BadRequest(c, "缺少fileId参数")
		return
	}

	// 递归搜索uploads目录查找匹配fileId前缀的STP/STEP文件
	uploadsDir := "./uploads"
	var foundPath string
	filepath.WalkDir(uploadsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || foundPath != "" {
			return nil
		}
		name := d.Name()
		if strings.HasPrefix(name, fileID) {
			lower := strings.ToLower(name)
			if strings.HasSuffix(lower, ".stp") || strings.HasSuffix(lower, ".step") {
				foundPath = path
			}
		}
		return nil
	})

	if foundPath == "" {
		NotFound(c, "未找到3D模型文件")
		return
	}

	// 获取绝对路径
	absPath, err := filepath.Abs(foundPath)
	if err != nil {
		InternalError(c, "解析文件路径失败")
		return
	}

	// 调用stp-thumbnail服务转换为STL
	reqBody, _ := json.Marshal(map[string]interface{}{
		"path":      absPath,
		"tolerance": 0.1,
	})
	resp, err := http.Post("http://127.0.0.1:5001/convert/stl", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		InternalError(c, "3D转换服务不可用")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		InternalError(c, "3D转换失败: "+string(body))
		return
	}

	c.Header("Content-Type", "application/sla")
	c.Header("Content-Disposition", "inline; filename=\"model.stl\"")
	io.Copy(c.Writer, resp.Body)
}
