vcl 4.1;

backend default {
    .host = "nginx";
    .port = "8080";
}

sub vcl_recv {

}

sub vcl_backend_response {

}

sub vcl_deliver {

}