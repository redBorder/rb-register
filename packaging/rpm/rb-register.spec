Name: rb-register
Version: %{__version}
Release: %{__release}%{?dist}

License: AGPL 3.0
URL: https://github.com/redBorder/rb-register
Source0: %{name}-%{version}.tar.gz

BuildRequires: go = 1.6.3
BuildRequires: glide rsync gcc git
BuildRequires:	rsync mlocate

Summary: rpm used to install rb-register in a redborder ng
Group:   Development/Libraries/Go

%description
%{summary}

%prep
%setup -qn %{name}-%{version}

%build

export GOPATH=${PWD}/gopath
export PATH=${GOPATH}:${PATH}

mkdir -p $GOPATH/src/github.com/redBorder/rb-register
rsync -az --exclude=packaging/ --exclude=resources/ --exclude=gopath/ ./ $GOPATH/src/github.com/redBorder/rb-register
cd $GOPATH/src/github.com/redBorder/rb-register
make

%install
mkdir -p %{buildroot}/usr/bin
mkdir -p %{buildroot}/etc/sysconfig
mkdir -p %{buildroot}/usr/lib/redborder/bin

mkdir -p %{buildroot}/usr/share/rb-register
mkdir -p %{buildroot}/etc/rb-register

export PARENT_BUILD=${PWD}
export GOPATH=${PWD}/gopath
export PATH=${GOPATH}:${PATH}
pushd $GOPATH/src/github.com/redBorder/rb-register
prefix=%{buildroot}/usr make install
popd
cp resources/bin/* %{buildroot}/usr/lib/redborder/bin
cp -f resources/files/rb-register.default %{buildroot}/etc/sysconfig/

install -D -m 0644 resources/systemd/rb-register.service %{buildroot}/usr/lib/systemd/system/rb-register.service

%clean
rm -rf %{buildroot}

%pre

%post
/usr/lib/redborder/bin/rb_rubywrapper.sh -c
[ ! -f /etc/sysconfig/rb-register ] && cp /etc/sysconfig/rb-register.default /etc/sysconfig/rb-register
systemctl daemon-reload

%files
%defattr(0755,root,root)
/usr/bin/rb_register
%defattr(644,root,root)
/usr/lib/systemd/system/rb-register.service
/etc/sysconfig/rb-register.default
%defattr(755,root,root)
/usr/lib/redborder/bin/rb_register_url.sh

%doc

%changelog
* Fri Nov 26 2021 Javier Rodriguez Gomez <javiercrg@redborder.com> - 0.0.1
- First spec version
