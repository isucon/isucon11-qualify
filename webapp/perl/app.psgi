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
    enable 'Session::Cookie',
        session_key => $ENV{SESSION_KEY} // 'isucondition',
        expires     => 3600,
        secret      => 'tagomoris';
    enable 'Static',
        path => qr!^/assets/!,
        root => $root_dir . '/../public/';
    $app;
};
