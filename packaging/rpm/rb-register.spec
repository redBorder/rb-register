Name:    rb-register
Version: %{__version}
Release: %{__release}%{?dist}

License: GNU AGPLv3
URL: https://github.com/redBorder/rb-register
Source0: %{name}-%{version}.tar.gz

BuildRequires: go = 1.6.3
BuildRequires: glide rsync gcc git
BuildRequires:	rsync mlocate pkgconfig
BuildRequires: librd-devel = 0.1.0
#BuildRequires: librdkafka-devel = 0.9.1

Requires: librd0 librdkafka1

Summary: Sensors to be registered byt the redborder manager
Group:   Development/Libraries/Go

%description
%{summary}

%prep
%setup -qn %{name}-%{version}

%build

git clone --branch v0.9.2 https://github.com/edenhill/librdkafka.git /tmp/librdkafka-v0.9.2
cd /tmp/librdkafka-v0.9.2
./configure --prefix=/usr --sbindir=/usr/bin --exec-prefix=/usr && make
make install
cd -
ldconfig
export PKG_CONFIG_PATH=/usr/lib/pkgconfig
export GOPATH=${PWD}/gopath
export PATH=${GOPATH}:${PATH}
mkdir -p $GOPATH/src/github.com/redBorder/rb-register
rsync -az --exclude=gopath/ ./ $GOPATH/src/github.com/redBorder/rb-register
cd $GOPATH/src/github.com/redBorder/rb-register
make

%install
mkdir -p %{buildroot}/usr/lib/redborder/bin
export PARENT_BUILD=${PWD}
export GOPATH=${PWD}/gopath
export PATH=${GOPATH}:${PATH}
export PKG_CONFIG_PATH=/usr/lib64/pkgconfig
cd $GOPATH/src/github.com/redBorder/rb-register
mkdir -p %{buildroot}/usr/bin
prefix=%{buildroot}/usr PKG_CONFIG_PATH=/usr/lib/pkgconfig/ make install
mkdir -p %{buildroot}/usr/share/rb-register
mkdir -p %{buildroot}/etc/rb-register
cp resources/bin/* %{buildroot}/usr/lib/redborder/bin
chmod 0755 %{buildroot}/usr/lib/redborder/bin/*
install -D -m 644 rb-register.service %{buildroot}/usr/lib/systemd/system/rb-register.service

%clean
rm -rf %{buildroot}

%pre

%post -p /sbin/ldconfig
%postun -p /sbin/ldconfig
systemctl daemon-reload

%files
%defattr(755,root,root)
/usr/lib/redborder/bin
/usr/bin/rb-register
%defattr(644,root,root)
/usr/lib/systemd/system/rb-register.service

%changelog
* Mon Mar 21 2022 Miguel Negrón <manegron@redborder.com> & David Vanhoucke <dvanhoucke@redborder.com> - 1.0.0-1
- first spec version
