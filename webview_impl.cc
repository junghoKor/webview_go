// webview_impl.cc

#define WEBVIEW_IMPLEMENTATION
#include "webview.h"

// OS별 헤더 포함 및 기능 구현
#ifdef _WIN32
    // [윈도우용]
    #include <windows.h>
#else
    // [리눅스용]
    #include <gtk/gtk.h>
#endif

// 기존 Show 함수
extern "C" void CgoWebViewShow(webview_t w) {
    auto *d = static_cast<webview::webview *>(w);

#ifdef _WIN32
    HWND hwnd = (HWND)d->window();
    ShowWindow(hwnd, SW_SHOW);
    UpdateWindow(hwnd);
#else
    GtkWidget *window = (GtkWidget *)d->window();
    gtk_widget_show_all(window);
#endif
}

// ★ [추가됨] 위치(x,y)와 크기(w,h)를 동시에 설정하는 함수
extern "C" void CgoWebViewSetBounds(webview_t w, int x, int y, int width, int height, int hint) {
    auto *d = static_cast<webview::webview *>(w);

    // 1. 기존 set_size를 먼저 호출하여 크기(Width, Height)와 힌트(Fixed, Min, Max)를 적용합니다.
    // webview.h 내부적으로 Windows는 SWP_NOMOVE를 사용하여 위치를 건드리지 않고,
    // GTK는 gtk_window_resize만 호출하므로 위치 이동 로직과 충돌하지 않습니다.
    d->set_size(width, height, (webview::webview_hint_t)hint);

    // 2. 위치(X, Y) 이동 처리
#ifdef _WIN32
    HWND hwnd = (HWND)d->window();
    // SWP_NOSIZE: 크기는 변경하지 않음 (위에서 처리했으므로)
    // SWP_NOZORDER: Z순서 변경 안 함
    // SWP_NOACTIVATE: 창을 활성화하여 포커스를 뺏지 않음
    SetWindowPos(hwnd, NULL, x, y, 0, 0, SWP_NOSIZE | SWP_NOZORDER | SWP_NOACTIVATE);
#else
    GtkWidget *window = (GtkWidget *)d->window();
    gtk_window_move(GTK_WINDOW(window), x, y);
#endif
}