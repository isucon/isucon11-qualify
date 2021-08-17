use FindBin;
use lib "$FindBin::Bin/extlib/lib/perl5";
use lib "$FindBin::Bin/lib";
use File::Basename;
use Plack::Builder;
use IsuCondition::Web;

my $root_dir = File::Basename::dirname(__FILE__);

my $app = IsuCondition::Web->psgi($root_dir);
builder {
    enable 'ReverseProxy';
    enable 'Static',
        path => qr!^/assets/!,
        root => $root_dir . '/../public/';
    enable 'Session::Cookie',
        session_key => 'isucondition_perl',
        expires     => 3600,
        secret      => $ENV{SESSION_KEY} || 'isucondition';
    $app;
};
