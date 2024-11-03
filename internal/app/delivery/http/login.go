package http

//
// type LoginUsecase interface {
// 	Login(ctx context.Context, req httpModel.Login) error
// }
//
// func (d *delivery) Login(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	var req httpModel.Login
// 	err := UnmarshalRequest(r.Body, &req)
// 	if err != nil {
// 		logger.Error(err.Error())
// 		ResponseWithJson(w, http.StatusBadRequest, Error{Msg: err.Error()})
// 		return
// 	}
//
// 	err = d.loginUsecase.Login(ctx, req)
// 	if err != nil {
// 		logger.Error(err.Error())
// 		ResponseWithJson(w, http.StatusInternalServerError, Error{Msg: err.Error()})
// 		return
// 	}
//
// 	ResponseWithJson(w, http.StatusOK, nil)
// }
