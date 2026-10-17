package web

import (
	"bytes"
	"fmt"
	"mime"
	"net/http"

	"github.com/samber/lo"
	"github.com/xuri/excelize/v2"

	"quizon/internal/app/usecase"
	"quizon/internal/utils"
)

const xlsxContentType string = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

var teamsHeader = []any{
	"№", "Команда", "Telegram", "ID команды", "Капитан", "Группа", "Телефон", "Участников",
	"Время регистрации",
}

func (wb *Web) exportTeams(w http.ResponseWriter, r *http.Request) {
	view, ok := wb.listTeams(w, r)
	if !ok {
		return
	}

	buf, err := teamsXLSX(view)
	if err != nil {
		wb.serverError(w, r, err)
		return
	}

	name := fmt.Sprintf("Команды игры %d.xlsx", view.Game.ID)
	w.Header().Set("Content-Type", xlsxContentType)
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": name}))
	w.WriteHeader(http.StatusOK)
	_, _ = buf.WriteTo(w)
}

func teamsXLSX(view teamsView) (*bytes.Buffer, error) {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	const sheet = "Команды"
	err := f.SetSheetName("Sheet1", sheet)
	if err != nil {
		return nil, fmt.Errorf("can't rename sheet: %w", err)
	}

	err = f.SetSheetRow(sheet, "A1", &teamsHeader)
	if err != nil {
		return nil, fmt.Errorf("can't write header: %w", err)
	}

	activeTeams := lo.Filter(view.Teams, func(r usecase.Registration, _ int) bool {
		return r.DeletedAt == nil
	})
	// в выгрузку идут только оставшиеся в игре команды
	for i, t := range activeTeams {
		row := []any{
			t.Number,
			t.TeamName,
			t.Telegram,
			lo.FromPtr(t.TeamID),
			t.CaptainName,
			lo.FromPtr(t.GroupName),
			t.Phone,
			t.PlayersAmount,
			t.CreatedAt.In(utils.LocMsk).Format("02.01.2006 15:04:05"),
		}

		err = f.SetSheetRow(sheet, fmt.Sprintf("A%d", i+2), &row)
		if err != nil {
			return nil, fmt.Errorf("can't write row: %w", err)
		}
	}

	err = f.SetColWidth(sheet, "A", "I", 18)
	if err != nil {
		return nil, fmt.Errorf("can't set column width: %w", err)
	}

	var buf bytes.Buffer
	err = f.Write(&buf)
	if err != nil {
		return nil, fmt.Errorf("can't write xlsx: %w", err)
	}

	return &buf, nil
}
