// stockapp/webview/webview_impl.cc
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

extern "C" void CgoWebViewShow(webview_t w) {
    auto *d = static_cast<webview::webview *>(w);

#ifdef _WIN32
    // 1. 윈도우(Windows)일 때
    HWND hwnd = (HWND)d->window();
    ShowWindow(hwnd, SW_SHOW);
    UpdateWindow(hwnd);
#else
    // 2. 리눅스(Linux)일 때
    // 리눅스에서는 윈도우 핸들이 'GtkWidget*' 포인터입니다.
    GtkWidget *window = (GtkWidget *)d->window();
    gtk_widget_show_all(window);
#endif
}