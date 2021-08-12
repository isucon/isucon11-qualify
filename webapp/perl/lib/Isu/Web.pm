package Isu::Web;
use v5.34;
use warnings;
use utf8;

use experimental qw(isa try);

use Kossy;

use DBIx::Sunny;
use File::Spec;
use HTTP::Status qw/:constants/;
use Log::Minimal;
use JSON::MaybeXS;
use Cpanel::JSON::XS::Type;

local $Log::Minimal::LOG_LEVEL = "DEBUG";

my $MYSQL_CONNECTION_DATA = {
    host     => $ENV{MYSQL_HOST}     // '127.0.0.1',
    port     => $ENV{MYSQL_PORT}     // '3306',
    user     => $ENV{MYSQL_USER}     // 'isucon',
    dbname   => $ENV{MYSQL_DATABASE} // 'isu',
    password => $ENV{MYSQL_PASS}     // 'isucon',
};

use constant InitializeResponse => {
    language => JSON_TYPE_STRING
};

my $_JSON = JSON::MaybeXS->new()->allow_blessed(1)->convert_blessed(1)->ascii(1);

# Initialize
post '/initialize' => \&initialize;

sub initialize {
    my ($self, $c)  = @_;

    my $sql_dir = File::Spec->catfile($self->root_dir, '..', 'mysql', 'db');
    my @path = (
        File::Spec->catfile($sql_dir, '0_Schema.sql'),
    );
    for my $p (@path) {
        my $cmd = sprintf("mysql -h %s -u %s -p%s -P %s %s < %s",
            $MYSQL_CONNECTION_DATA->{host},
            $MYSQL_CONNECTION_DATA->{user},
            $MYSQL_CONNECTION_DATA->{password},
            $MYSQL_CONNECTION_DATA->{port},
            $MYSQL_CONNECTION_DATA->{dbname},
            $p
        );
        if (my $e = system($cmd)) {
            infof('Initialize script error : %s , %s', $e, $!);
            return $self->res_no_content($c, HTTP_INTERNAL_SERVER_ERROR);
        }
    }

    $self->res_json($c, {
        language => "perl",
    }, InitializeResponse);
};

filter 'allow_json_request' => sub {
    my $app = shift;
    return sub {
        my ($self, $c) = @_;
        $c->env->{'kossy.request.parse_json_body'} = 1;
        $app->($self, $c);
    };
};

sub dbh {
    my $self = shift;
    $self->{_dbh} ||= do {
        my ($host, $port, $user, $dbname, $password) = $MYSQL_CONNECTION_DATA->@{qw/host port user dbname password/};
        my $dsn = "dbi:mysql:database=$dbname;host=$host;port=$port";
        DBIx::Sunny->connect($dsn, $user, $password, {
            mysql_enable_utf8mb4 => 1,
            mysql_auto_reconnect => 1,
            Callbacks => {
                connected => sub {
                    my $dbh = shift;
                    # XXX $dbh->do('SET SESSION sql_mode="STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION"');
                    return;
                },
            },
        });
    };
}

# send empty body with status code
sub res_no_content {
    my ($self, $c, $status) = @_;
    $c->res->headers->remove_content_headers;
    $c->res->content_length(0);
    $c->res->code($status);
    $c->res;
}

# render_json with json spec
# XXX: $json_specを指定できるようにKossy::Conection#render_jsonを調整
sub res_json {
    my ($self, $c, $obj, $json_spec) = @_;

    my $body = $_JSON->encode($obj, $json_spec);
    $body = $c->escape_json($body);

    if ( ( $c->req->env->{'HTTP_USER_AGENT'} || '' ) =~ m/Safari/ ) {
        $body = "\xEF\xBB\xBF" . $body;
    }

    $c->res->status( 200 );
    $c->res->content_type('application/json; charset=UTF-8');
    $c->res->header( 'X-Content-Type-Options' => 'nosniff' ); # defense from XSS
    $c->res->body( $body );
    $c->res;
}

1
