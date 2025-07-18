while (1) {
    handle_request();
    last if $shutdown;
}
