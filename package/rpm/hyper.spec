Summary:            Hyper is a VM based docker runtime
Name:               hyper
Version:            0.4
Release:            1%{?dist}
License:            Apache License, Version 2.0
Group:              System Environment/Base
# The source for this package was pulled from upstream's git repo. Use the
# following commands to generate the tarball:
#  git archive --format=tar.gz master > hyper-%{version}.tar.gz
Source0:            %{name}-%{version}.tar.gz
# and the https://github.com/hyperhq/runv.git
#  git archive --format=tar.gz master > runv-%{version}.tar.gz
Source1:            runv-%{version}.tar.gz
URL:                https://hyper.sh/
ExclusiveArch:      x86_64
Requires:           device-mapper,sqlite
BuildRequires:      device-mapper-devel,pcre-devel,libsepol-devel,libselinux-devel,systemd-devel
BuildRequires:      sqlite-devel
BuildRequires:      libuuid-devel,xen-devel

%description
Hyper is a VM based docker engine, it start a container image in
VM without a full guest OS

%prep
mkdir -p %{_builddir}/src/github.com/hyperhq/hyper
mkdir -p %{_builddir}/src/github.com/hyperhq/runv
tar -C %{_builddir}/src/github.com/hyperhq/hyper -xvf %SOURCE0
tar -C %{_builddir}/src/github.com/hyperhq/runv -xvf %SOURCE1

%build
cd %{_builddir}/src/github.com/hyperhq/hyper
export GOPATH=%{_builddir}
./autogen.sh
./configure
make %{?_smp_mflags}

%install
mkdir -p %{buildroot}%{_bindir}
mkdir -p %{buildroot}%{_sysconfdir}
cp %{_builddir}/src/github.com/hyperhq/hyper/{hyper,hyperd} %{buildroot}%{_bindir}
cp -a %{_builddir}/src/github.com/hyperhq/hyper/package/dist/etc/hyper %{buildroot}%{_sysconfdir}

%clean
rm -rf %{buildroot}

%files
%{_bindir}/*
%{_sysconfdir}/*

%changelog
* Sat Nov 21 2015 Xu Wang <xu@hyper.sh> - 0.4-1
- Initial rpm packaging