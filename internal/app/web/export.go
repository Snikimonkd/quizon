package web

import (
	"bytes"
	"fmt"
	"mime"
	"net/http"

	"quizon/internal/utils"

	"github.com/samber/lo"
	"github.com/xuri/excelize/v2"
)

const xlsxContentType string = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

var teamsHeader = []any{
	"№", "Состав", "Команда", "Telegram", "ID команды", "Капитан", "Группа", "Телефон", "Участников",
	"Время регистрации",
}

func (wb *Web) exportTeams(w http.ResponseWriter, r *http.Request) {
	view, ok := wb.listTeams(w, r)
	if !ok {
		return
	}

	buf, err := teamsXLSX(view)
	if err != nil {
		wb.serverError(w, err)
		return
	}

	name := fmt.Sprintf("Команды игры %d.xlsx", view.Game.ID)
	w.Header().Set("Content-Type", xlsxContentType)
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": name}))
	w.WriteHeader(http.StatusOK)
	_, _ = buf.WriteTo(w)
}

// squad - основной состав или резерв по порядковому номеру регистрации
func squad(number, mainAmount int64) string {
	if number > mainAmount {
		return "резерв"
	}

	return "основа"
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

	for i, t := range view.Teams {
		row := []any{
			t.Number,
			squad(t.Number, view.MainAmount),
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

	err = f.SetColWidth(sheet, "A", "J", 18)
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
