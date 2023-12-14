Name: rb-register
Version: %{__version}
Release: %{__release}%{?dist}

License: AGPL 3.0
URL: https://github.com/redBorder/rb-register
Source0: %{name}-%{version}.tar.gz

BuildRequires: go rsync gcc git

Summary: rpm used to install rb-register in a redborder ng
Group:   Development/Libraries/Go

%global debug_package %{nil}

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
install -D -m 0644 resources/systemd/rb-register.service %{buildroot}/usr/lib/systemd/system/rb-register.service


%clean
rm -rf %{buildroot}

%pre

%post
systemctl daemon-reload
mkdir -p /var/log/rb-register
[ -f /usr/lib/redborder/bin/rb_rubywrapper.sh ] && /usr/lib/redborder/bin/rb_rubywrapper.sh -c


%files
%defattr(0755,root,root)
/usr/bin/rb_register
%defattr(644,root,root)
/usr/lib/systemd/system/rb-register.service
%defattr(755,root,root)
/usr/lib/redborder/bin/rb_register_url.sh
/usr/lib/redborder/bin/rb_register_finish.sh

%doc

%changelog
* Thu Dec 14 2023 Miguel Álvarez <malvarez@redborder.com> - 2.0.1-1
- add cgroups call
* Wed Oct 04 2023 David Vanhoucke <dvanhoucke@redborder.com> - 2.0.0-1
- adapt for go mod
* Wed Mar 30 2022 Miguel Negron <manegron@redborder.com> - 1.1.10
- Make rb-register generic
* Fri Nov 26 2021 Javier Rodriguez Gomez <javiercrg@redborder.com> - 0.0.1
- First spec version
