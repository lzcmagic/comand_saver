package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

func initDB() *sql.DB {
	// 获取用户主目录
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("获取用户主目录失败: %v", err))
	}

	// 使用 filepath.Join 来处理跨平台的路径
	dbDir := filepath.Join(home, ".command_saver")
	dbPath := filepath.Join(dbDir, "commands.db")

	// 检查目录是否存在，如果不存在则创建
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		fmt.Printf("数据库目录不存在，正在创建: %s\n", dbDir)
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			panic(fmt.Sprintf("创建数据库目录失败: %v", err))
		}
	}

	// 打开或创建数据库
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		panic(fmt.Sprintf("打开数据库失败: %v", err))
	}

	// 创建表（如果不存在）
	createTableSQL := `
	PRAGMA foreign_keys = ON;
	PRAGMA encoding = 'UTF-8';
	
	CREATE TABLE IF NOT EXISTS command_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		command TEXT NOT NULL COLLATE NOCASE,
		description TEXT COLLATE NOCASE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err = db.Exec(createTableSQL); err != nil {
		panic(fmt.Sprintf("创建数据表失败: %v", err))
	}

	return db
}

func saveCommand(db *sql.DB, command, description string) {
	// 清理输入
	command = strings.TrimSpace(command)
	description = strings.TrimSpace(description)

	stmt, err := db.Prepare("INSERT INTO command_history(command, description, created_at) VALUES(?, ?, datetime('now', 'localtime'))")
	if err != nil {
		fmt.Println("准备SQL语句时出错:", err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(command, description)
	if err != nil {
		fmt.Println("保存命令时出错:", err)
		return
	}

	if description == "default" {
		fmt.Printf("命令已保存：%s\n", command)
	} else {
		fmt.Printf("命令已保存：%s (描述: %s)\n", command, description)
	}
}

func getLastCommand() string {
	// Windows 平台暂不支持自动获取上一条命令
	if runtime.GOOS == "windows" {
		fmt.Println("Windows 平台暂不支持自动获取上一条命令，请使用 -y 参数手动保存命令")
		return ""
	}

	// 检测当前的 shell
	shell := os.Getenv("SHELL")
	if shell == "" {
		fmt.Println("无法检测到当前shell类型")
		return ""
	}

	var histFile string
	var isZsh bool

	// 判断shell类型
	if strings.Contains(shell, "zsh") {
		home, _ := os.UserHomeDir()
		histFile = filepath.Join(home, ".zsh_history")
		isZsh = true
	} else if strings.Contains(shell, "bash") {
		home, _ := os.UserHomeDir()
		histFile = filepath.Join(home, ".bash_history")
		isZsh = false
	} else {
		fmt.Printf("不支持的shell类型: %s\n", shell)
		return ""
	}

	// 检查历史文件是否存在
	if _, err := os.Stat(histFile); os.IsNotExist(err) {
		fmt.Printf("历史文件不存在: %s\n", histFile)
		return ""
	}

	// 直接读取历史文件
	content, err := os.ReadFile(histFile)
	if err != nil {
		fmt.Printf("读取历史文件出错: %v\n", err)
		return ""
	}

	// 将内容转换为字符串并按行分割
	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 {
		fmt.Println("历史文件为空")
		return ""
	}

	// 从后向前遍历，查找最后一条有效命令
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		// 处理 zsh 特殊格式
		if isZsh {
			// zsh历史格式可能是: ": 时间戳:0;命令"
			if strings.Contains(line, ":0;") {
				parts := strings.SplitN(line, ":0;", 2)
				if len(parts) >= 2 {
					line = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(line, ";") {
				parts := strings.SplitN(line, ";", 2)
				if len(parts) >= 2 {
					line = strings.TrimSpace(parts[1])
				}
			}
		}

		// 排除当前程序的命令和空命令
		if line == "" ||
			strings.HasPrefix(line, "./cs") ||
			strings.HasPrefix(line, "cs ") ||
			strings.HasPrefix(line, "go run main.go") {
			continue
		}

		// 解析命令
		cmdParts := strings.Fields(line)
		if len(cmdParts) == 0 {
			continue
		}

		// 验证命令
		cmdName := cmdParts[0]
		if strings.Contains(cmdName, "/") || strings.HasPrefix(cmdName, "./") {
			// 处理相对路径或绝对路径
			var fullPath string
			if strings.HasPrefix(cmdName, "./") {
				pwd, err := os.Getwd()
				if err != nil {
					fmt.Printf("获取当前工作目录失败: %v\n", err)
					continue
				}
				fullPath = filepath.Join(pwd, cmdName[2:])
			} else {
				fullPath = cmdName
			}

			if _, err := os.Stat(fullPath); err != nil {
				continue
			}
		} else {
			// 使用 which 命令检查系统命令
			cmd := exec.Command("which", cmdName)
			if err := cmd.Run(); err != nil {
				continue
			}
		}

		fmt.Printf("找到最后一条命令: %s\n", line)
		return line
	}

	fmt.Println("未找到有效的历史命令")
	return ""
}

func listCommands(db *sql.DB) {
	rows, err := db.Query(`
		SELECT id, command, description, created_at 
		FROM command_history 
		ORDER BY created_at DESC
	`)
	if err != nil {
		fmt.Println("查询数据库时出错:", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n保存的命令历史:")
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Printf("%-6s | %-30s | %-30s | %s\n", "ID", "时间", "命令", "描述")
	fmt.Println("--------------------------------------------------------------------------------")

	for rows.Next() {
		var id int
		var command, description string
		var createdAt time.Time
		err := rows.Scan(&id, &command, &description, &createdAt)
		if err != nil {
			fmt.Println("读取数据时出错:", err)
			continue
		}
		timeStr := createdAt.Format("2006-01-02 15:04:05")
		fmt.Printf("%-6d | %-30s | %-30s | %s\n", id, timeStr, command, description)
	}
	fmt.Println("--------------------------------------------------------------------------------")
}

// 添加新的函数来显示天分组的命令
func listCommandsByDay(db *sql.DB) {
	// 查询近7天的命令，按天分组
	rows, err := db.Query(`
		WITH RECURSIVE dates(date) AS (
			SELECT date('now', 'localtime', '-6 days')
			UNION ALL
			SELECT date(date, '+1 day')
			FROM dates
			WHERE date < date('now', 'localtime')
		)
		SELECT 
			dates.date as day,
			GROUP_CONCAT(id) as ids,
			GROUP_CONCAT(command) as commands,
			GROUP_CONCAT(description) as descriptions,
			GROUP_CONCAT(created_at) as times
		FROM dates 
		LEFT JOIN command_history ON date(command_history.created_at) = dates.date
		GROUP BY dates.date
		ORDER BY dates.date DESC
	`)
	if err != nil {
		fmt.Println("查询数据库时出错:", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n最近7天的命令历史:")

	for rows.Next() {
		var day string
		var ids, commands, descriptions, times sql.NullString

		err := rows.Scan(&day, &ids, &commands, &descriptions, &times)
		if err != nil {
			fmt.Println("读取数据时出错:", err)
			continue
		}

		// 打印日期分隔线
		fmt.Printf("\n=== %s ===\n", day)
		fmt.Println("--------------------------------------------------------------------------------")
		fmt.Printf("%-6s | %-30s | %-30s | %s\n", "ID", "时间", "命令", "描述")
		fmt.Println("--------------------------------------------------------------------------------")

		// 如果这一天没有命令，继续下一天
		if !ids.Valid || ids.String == "" {
			fmt.Println("(没有记录)")
			continue
		}

		// 分割每一天的数据
		idList := strings.Split(ids.String, ",")
		cmdList := strings.Split(commands.String, ",")
		descList := strings.Split(descriptions.String, ",")
		timeList := strings.Split(times.String, ",")

		// 确保所有切片长度一致
		length := len(idList)
		for i := 0; i < length; i++ {
			id := idList[i]
			cmd := cmdList[i]
			desc := descList[i]
			if desc == "" {
				desc = "-"
			}

			// 解析并格式化时间
			t, err := time.Parse("2006-01-02 15:04:05", strings.Split(timeList[i], ".")[0])
			if err != nil {
				fmt.Printf("解析时间出错: %v\n", err)
				continue
			}
			timeStr := t.Format("15:04:05")

			fmt.Printf("%-6s | %-30s | %-30s | %s\n", id, timeStr, cmd, desc)
		}
		fmt.Println("--------------------------------------------------------------------------------")
	}
}

func cleanDatabase() {
	// 获取数据库文件路径
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("获取用户主目录失败: %v\n", err)
		return
	}
	dbPath := filepath.Join(home, ".command_saver", "commands.db")

	// 检查文件是否存在
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Println("数据库文件不存在")
		return
	}

	// 询问用户确认
	fmt.Print("警告：此操作将删除所有保存的命令历史记录，确定要继续吗？(y/N): ")
	var response string
	fmt.Scanln(&response)

	// 检查用户响应
	response = strings.ToLower(strings.TrimSpace(response))
	if response != "y" && response != "yes" {
		fmt.Println("操作已取消")
		return
	}

	// 删除数据库文件
	err = os.Remove(dbPath)
	if err != nil {
		fmt.Printf("删除数据库文件失败: %v\n", err)
		return
	}

	fmt.Println("数据库已清除")
}

func deleteCommand(db *sql.DB, id int) {
	// 首先检查记录是否存在
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM command_history WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		fmt.Printf("检查记录时出错: %v\n", err)
		return
	}

	if !exists {
		fmt.Printf("未找到ID为 %d 的记录\n", id)
		return
	}

	// 执行删除操作
	result, err := db.Exec("DELETE FROM command_history WHERE id = ?", id)
	if err != nil {
		fmt.Printf("删除记录时出错: %v\n", err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		fmt.Printf("成功删除ID为 %d 的记录\n", id)
	}
}

// 添加新的结构体用于JSON导出
type CommandRecord struct {
	ID          int       `json:"id"`
	Command     string    `json:"command"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

func exportToJSON(db *sql.DB, filename string) {
	// 如果文件名为空，使用默认文件名
	if filename == "" {
		timestamp := time.Now().Format("20060102_150405")
		filename = fmt.Sprintf("bak%s.json", timestamp)
	}

	// 获取数据库目录
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("获取用户主目录失败: %v\n", err)
		return
	}
	dbDir := filepath.Join(home, ".command_saver")
	outputPath := filepath.Join(dbDir, filename)

	// 查询所有命令记录
	rows, err := db.Query(`
		SELECT id, command, description, created_at 
		FROM command_history 
		ORDER BY created_at DESC
	`)
	if err != nil {
		fmt.Printf("查询数据库失败: %v\n", err)
		return
	}
	defer rows.Close()

	// 读取所有记录
	var records []CommandRecord
	for rows.Next() {
		var record CommandRecord
		err := rows.Scan(&record.ID, &record.Command, &record.Description, &record.CreatedAt)
		if err != nil {
			fmt.Printf("读取记录失败: %v\n", err)
			return
		}
		records = append(records, record)
	}

	// 转换为JSON
	jsonData, err := json.MarshalIndent(records, "", "    ")
	if err != nil {
		fmt.Printf("转换JSON失败: %v\n", err)
		return
	}

	// 写入文件
	err = os.WriteFile(outputPath, jsonData, 0644)
	if err != nil {
		fmt.Printf("写入文件失败: %v\n", err)
		return
	}

	fmt.Printf("命令历史已导出到: %s\n", outputPath)
}

func importFromJSON(db *sql.DB, filename string) {
	if filename == "" {
		fmt.Println("错误: 请指定要导入的JSON文件路径")
		fmt.Println("使用方法: cs -i \"<备份文件路径>\"")
		return
	}

	// 读取JSON文件
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		return
	}

	// 解析JSON数据
	var records []CommandRecord
	if err := json.Unmarshal(jsonData, &records); err != nil {
		fmt.Printf("解析JSON失败: %v\n", err)
		return
	}

	if len(records) == 0 {
		fmt.Println("文件中没有找到任何命令记录")
		return
	}

	// 准备插入语句
	stmt, err := db.Prepare("INSERT INTO command_history(command, description, created_at) VALUES(?, ?, ?)")
	if err != nil {
		fmt.Printf("准备SQL语句失败: %v\n", err)
		return
	}
	defer stmt.Close()

	// 开始导入
	fmt.Printf("开始导入 %d 条命令记录...\n", len(records))
	successCount := 0

	for _, record := range records {
		_, err := stmt.Exec(record.Command, record.Description, record.CreatedAt)
		if err != nil {
			fmt.Printf("导入命令 '%s' 失败: %v\n", record.Command, err)
			continue
		}
		successCount++
	}

	fmt.Printf("导入完成: 成功导入 %d/%d 条记录\n", successCount, len(records))
}

func showHelp() {
	fmt.Println("使用方法:")
	fmt.Println("  cs [描述]            保存上一条执行的命令")
	fmt.Println("  cs -l               列出所有保存的命令")
	fmt.Println("  cs -d               按天显示最近7天的命令")
	fmt.Println("  cs -y <命令> [描述]  直接保存指定的命令")
	fmt.Println("  cs -rm <id>         删除指定ID的命令记录")
	fmt.Println("  cs -o [文件名]       导出命令历史到JSON文件")
	fmt.Println("  cs -i <文件名>       从JSON文件导入命令历史")
	fmt.Println("  cs -h               显示帮助信息")
	fmt.Println("  cs -c               清理数据库")
}

func main() {
	// 首先处理帮助命令
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		showHelp()
		return
	}

	// 处理其他命令行参数
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-l":
			db := initDB()
			defer db.Close()
			listCommands(db)
			return
		case "-d":
			db := initDB()
			defer db.Close()
			listCommandsByDay(db)
			return
		case "-c":
			cleanDatabase()
			return
		case "-rm":
			if len(os.Args) != 3 {
				fmt.Println("错误: 使用 -rm 参数时必须提供要删除的记录ID")
				return
			}
			id := 0
			_, err := fmt.Sscanf(os.Args[2], "%d", &id)
			if err != nil || id <= 0 {
				fmt.Println("错误: ID必须是一个有效的正整数")
				return
			}
			db := initDB()
			defer db.Close()
			deleteCommand(db, id)
			return
		case "-y":
			if len(os.Args) < 3 {
				fmt.Println("错误: 使用 -y 参数时必须提供要保存的命令")
				return
			}

			// 获取命令和描述
			command := strings.TrimSpace(os.Args[2])
			description := "default"

			// 如果有第三个参数，则作为描���
			if len(os.Args) > 3 {
				description = strings.TrimSpace(strings.Join(os.Args[3:], " "))
			}

			// 去除命令和描述中的引号
			command = strings.Trim(command, "\"'")
			description = strings.Trim(description, "\"'")

			db := initDB()
			defer db.Close()
			saveCommand(db, command, description)
			return
		case "-o":
			db := initDB()
			defer db.Close()
			var filename string
			if len(os.Args) > 2 {
				filename = strings.Trim(os.Args[2], "\"'")
			}
			exportToJSON(db, filename)
			return
		case "-i":
			db := initDB()
			defer db.Close()
			var filename string
			if len(os.Args) > 2 {
				filename = strings.Trim(os.Args[2], "\"'")
			}
			importFromJSON(db, filename)
			return
		}
	}

	// 初始化数据库
	db := initDB()
	defer db.Close()

	// 获取上一条命令
	lastCommand := getLastCommand()
	if lastCommand == "" {
		// 如果已经打印了具体错误信息，就不再显示通用错误
		return
	}

	// 获取可选的描述
	var description string
	if len(os.Args) > 1 {
		description = strings.TrimSpace(strings.Join(os.Args[1:], " "))
	} else {
		description = "default"
	}

	// 保存到数据库
	saveCommand(db, lastCommand, description)
}
